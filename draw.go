package main

import (
	"fmt"
	"image"
)

type MapData struct {
	TileSize            int
	WallSize            int
	WallTop, WallBottom uint8   // Greyscale value of the wall
	Shadow, ShadowDepth uint8   // Shadow is weakest shadow alpha (default is ~20%), strongest shadow is ShadowDepth plus Shadow
	ShadowWidth         float64 // Falloff (1 means half shadow at one tile away, 2 means half shadow at 2 tiles away, etc)
	Light               uint8   // wall at LightAngle angle this (max) amplification
	LightAngle          int16   // in degrees
	Output              string  // Output filename (must end in .png)
	line                []string
	pic                 image.Image
}

func (m *MapData) defaults() {
	m.TileSize = 50
	m.WallSize = 6
	m.WallTop = 160
	m.WallBottom = 135
	m.Shadow = 50
	m.ShadowDepth = 65
	m.ShadowWidth = 0.2
	m.Light = 15
	m.LightAngle = 10
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
		for x, t := range row {
			switch t {
			case '#':
				// translucent black
				for ry := 0; ry < m.TileSize; ry++ {
					for rx := 0; rx < m.TileSize; rx++ {
						i.Pix[origin+rx*4+ry*i.Stride+3] = m.Shadow
					}
				}
			case '\\':
				// lower-left is solid
				before := m.Shadow
				after := uint8(255)
				if m.line[y][x+1] == '#' || m.line[y-1][x] == '#' ||
					m.line[y][x+1] == '\\' || m.line[y-1][x] == '\\' {
					// upper-right is solid
					before, after = after, before
				}
				for ry := 0; ry < m.TileSize; ry++ {
					for rx := 0; rx < m.TileSize; rx++ {
						val := after
						if rx <= ry {
							val = before
						}
						i.Pix[origin+rx*4+ry*i.Stride+3] = val
					}
				}
			case '/':
				// lower-right is solid
				before := m.Shadow
				after := uint8(255)
				if m.line[y][x-1] == '#' || m.line[y-1][x] == '#' ||
					m.line[y][x-1] == '/' || m.line[y-1][x] == '/' {
					// upper-left is solid
					before, after = after, before
				}
				for ry := 0; ry < m.TileSize; ry++ {
					for rx := 0; rx < m.TileSize; rx++ {
						val := after
						if m.TileSize-rx-1 <= ry {
							val = before
						}
						i.Pix[origin+rx*4+ry*i.Stride+3] = val
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
