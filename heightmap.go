package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"sync"
)

type Bounds struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
}

type TheaterConfig struct {
	HeightmapBounds Bounds
	HeightmapWidth  int
	HeightmapHeight int
}

type HeightmapNotFoundError struct{ Msg string }

func (e HeightmapNotFoundError) Error() string { return e.Msg }

type CoordinatesOutOfBoundsError struct{ Msg string }

func (e CoordinatesOutOfBoundsError) Error() string { return e.Msg }

type HeightmapReadError struct {
	Msg string
	Err error
}

func (e HeightmapReadError) Error() string {
	if e.Err == nil {
		return e.Msg
	}
	return e.Msg + ": " + e.Err.Error()
}
func (e HeightmapReadError) Unwrap() error { return e.Err }

type HeightmapReader struct {
	path string
	cfg  TheaterConfig

	mu   sync.Mutex
	file *os.File

	invSpanX float64
	invSpanY float64
}

func NewHeightmapReader(path string, cfg TheaterConfig) (*HeightmapReader, error) {
	if cfg.HeightmapWidth <= 0 || cfg.HeightmapHeight <= 0 {
		return nil, fmt.Errorf("invalid heightmap size %dx%d", cfg.HeightmapWidth, cfg.HeightmapHeight)
	}
	spanX := cfg.HeightmapBounds.MaxX - cfg.HeightmapBounds.MinX
	spanY := cfg.HeightmapBounds.MaxY - cfg.HeightmapBounds.MinY
	if spanX <= 0 || spanY <= 0 {
		return nil, fmt.Errorf("invalid heightmap bounds spanX=%v spanY=%v", spanX, spanY)
	}

	return &HeightmapReader{
		path:     path,
		cfg:      cfg,
		invSpanX: 1.0 / spanX,
		invSpanY: 1.0 / spanY,
	}, nil
}

func (r *HeightmapReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}

func (r *HeightmapReader) ensureOpen() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file != nil {
		return nil
	}

	f, err := os.Open(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return HeightmapNotFoundError{Msg: "heightmap file not found or cannot be accessed"}
		}
		return HeightmapReadError{Msg: "error opening heightmap file", Err: err}
	}
	r.file = f
	return nil
}

func (r *HeightmapReader) GetElevation(xPos, yPos float64) (float64, error) {
	if err := r.ensureOpen(); err != nil {
		return 0, err
	}

	b := r.cfg.HeightmapBounds
	if xPos < b.MinX || xPos > b.MaxX || yPos < b.MinY || yPos > b.MaxY {
		return 0, CoordinatesOutOfBoundsError{
			Msg: fmt.Sprintf(
				"coordinates (%.1f, %.1f) are outside valid bounds (%.1f-%.1f, %.1f-%.1f)",
				xPos, yPos, b.MinX, b.MaxX, b.MinY, b.MaxY,
			),
		}
	}

	xRatio := (xPos - b.MinX) * r.invSpanX
	yRatio := (yPos - b.MinY) * r.invSpanY

	w := r.cfg.HeightmapWidth
	h := r.cfg.HeightmapHeight

	px := int(xRatio * float64(w-1))
	py := int((1.0 - yRatio) * float64(h-1))

	// Clamp
	if px < 0 {
		px = 0
	} else if px >= w {
		px = w - 1
	}
	if py < 0 {
		py = 0
	} else if py >= h {
		py = h - 1
	}

	// Offset: index * 2 (uint16)
	index := py*w + px
	off := int64(index * 2)

	var buf [2]byte

	// ReaderAt: thread-safe Reads (os.File.ReadAt ist parallel nutzbar)
	// Wir greifen aber auf r.file unter Mutex zu, damit Close/Swap sauber bleibt.
	r.mu.Lock()
	f := r.file
	r.mu.Unlock()

	if f == nil {
		return 0, HeightmapReadError{Msg: "heightmap file is not open"}
	}

	n, err := f.ReadAt(buf[:], off)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, HeightmapReadError{Msg: "error reading heightmap file", Err: err}
	}
	if n != 2 {
		return 0, HeightmapReadError{Msg: fmt.Sprintf("expected 2 bytes, got %d bytes", n)}
	}

	rawElevation := binary.LittleEndian.Uint16(buf[:])

	return float64(rawElevation), nil
}

type ExportHeightmapOptions struct {
	DstSize       int     // e.g. 1024
	LittleEndian  bool    // most RAW heightmaps are little-endian
	FeetPerUnit   float64 // multiply raw sample by this to get feet
	ClampNegative bool    // if conversion yields <0, clamp to 0
}

// ExportHeightmap reads the entire RAW heightmap (uint16 samples),
// downsamples to DstSize x DstSize by taking the MAX height in each block,
// then writes an 8-bit grayscale PNG.
//
// Color mapping:
// - Each unit in the 8-bit grayscale represents 100 feet (0=0ft, 1=100ft, ..., 255=25500ft)
func ExportHeightmap(rawPath string, srcW, srcH int, outPath string, opt ExportHeightmapOptions) error {
	if opt.DstSize <= 0 {
		return fmt.Errorf("DstSize must be > 0")
	}
	if srcW <= 0 || srcH <= 0 {
		return fmt.Errorf("invalid source size %dx%d", srcW, srcH)
	}
	if opt.FeetPerUnit == 0 {
		return fmt.Errorf("FeetPerUnit must not be 0")
	}

	f, err := os.Open(rawPath)
	if err != nil {
		return err
	}
	defer f.Close()

	br := bufio.NewReaderSize(f, 1<<20)

	dst := opt.DstSize
	blockW := int(math.Ceil(float64(srcW) / float64(dst)))
	blockH := int(math.Ceil(float64(srcH) / float64(dst)))
	if blockW <= 0 {
		blockW = 1
	}
	if blockH <= 0 {
		blockH = 1
	}

	// Store per-block maxima in feet.
	blockMaxFeet := make([]uint32, dst*dst) // feet, clamped to >=0

	// Helper: read next uint16 sample (raw).
	readU16 := func() (uint16, error) {
		var b [2]byte
		_, err := io.ReadFull(br, b[:])
		if err != nil {
			return 0, err
		}
		if opt.LittleEndian {
			return binary.LittleEndian.Uint16(b[:]), nil
		}
		return binary.BigEndian.Uint16(b[:]), nil
	}

	// Stream the entire file once, compute per-block maxima.
	for y := 0; y < srcH; y++ {
		by := y / blockH
		if by >= dst {
			by = dst - 1
		}
		rowBase := by * dst

		for x := 0; x < srcW; x++ {
			raw, err := readU16()
			if err != nil {
				return fmt.Errorf("read sample at (%d,%d): %w", x, y, err)
			}

			feet := float64(raw) * opt.FeetPerUnit
			if opt.ClampNegative && feet < 0 {
				feet = 0
			}
			if feet < 0 {
				// If not clamping, still avoid wrapping in uint32.
				feet = 0
			}

			bx := x / blockW
			if bx >= dst {
				bx = dst - 1
			}

			idx := rowBase + bx
			v := uint32(math.Round(feet))
			if v > blockMaxFeet[idx] {
				blockMaxFeet[idx] = v
			}
		}
	}

	// Build 8-bit grayscale image: 1 unit = 100 ft.
	img := image.NewGray(image.Rect(0, 0, dst, dst))
	for y := 0; y < dst; y++ {
		for x := 0; x < dst; x++ {
			feet := blockMaxFeet[y*dst+x]
			val := float64(feet) / 100.0
			if val > 255 {
				val = 255
			}
			img.Pix[y*img.Stride+x] = uint8(math.Round(val))
		}
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	bw := bufio.NewWriterSize(out, 1<<20)
	if err := png.Encode(bw, img); err != nil {
		return err
	}
	return bw.Flush()
}
