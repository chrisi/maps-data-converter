package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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
