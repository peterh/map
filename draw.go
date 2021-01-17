package main

import (
	"fmt"
	"image"
)

type MapData struct {
	WallSize int
	TileSize int
	Output   string
	line     []string
	pic      image.Image
}

func (m *MapData) draw() {
	width := 0
	height := len(m.line) * m.TileSize
	if len(m.line) > 0 {
		width = len(m.line[0]) * m.TileSize
	}

	i := image.NewNRGBA(image.Rect(0, 0, width, height))
	m.pic = i

	for y, row := range m.line {
		origin := y * m.TileSize * i.Stride
		for _, t := range row {
			switch t {
			case '#':
				// translucent black
				for ry := 0; ry < m.TileSize; ry++ {
					for rx := 0; rx < m.TileSize; rx++ {
						i.Pix[origin+rx*4+ry*i.Stride+3] = 60
					}
				}
			default:
				fmt.Printf("unrecognized token %c\n", t)
				fallthrough
			case ' ':
				// opaque black
				for ry := 0; ry < m.TileSize; ry++ {
					for rx := 0; rx < m.TileSize; rx++ {
						i.Pix[origin+rx*4+ry*i.Stride+3] = 255
					}
				}
			}
			origin += m.TileSize * 4
		}
	}
}
