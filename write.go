package main

import (
	"errors"
	"image/png"
	"os"
	"strings"
)

func (m *MapData) write() error {
	if !strings.HasSuffix(m.Output, ".png") {
		return errors.New("Fatal: Only png supported so far")
	}
	f, err := os.Create(m.Output)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, m.pic)
}
