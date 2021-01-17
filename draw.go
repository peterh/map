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

	floor := make([][]bool, height)
	for i := range floor {
		floor[i] = make([]bool, width)
	}

	for y, row := range m.line {
		for x, t := range row {
			switch t {
			case '#':
				// translucent black
				for dy := 0; dy < m.TileSize; dy++ {
					for dx := 0; dx < m.TileSize; dx++ {
						floor[y*m.TileSize+dy][x*m.TileSize+dx] = true
					}
				}
			case '\\':
				// lower-left is solid
				before := true
				after := false
				if m.line[y][x+1] == '#' || m.line[y-1][x] == '#' {
					// strong rule: upper-right is solid
					before, after = after, before
				} else if (m.line[y][x+1] != ' ' || m.line[y-1][x] != ' ') &&
					m.line[y][x-1] != '#' && m.line[y+1][x] != '#' {
					// weak rule: upper-right is solid if we see upper-right and don't see solid below/left
					before, after = after, before
				}
				for dy := 0; dy < m.TileSize; dy++ {
					for dx := 0; dx < m.TileSize; dx++ {
						val := after
						if dx <= dy {
							val = before
						}
						floor[y*m.TileSize+dy][x*m.TileSize+dx] = val
					}
				}
			case '/':
				// lower-right is solid
				before := true
				after := false
				if m.line[y][x-1] == '#' || m.line[y-1][x] == '#' {
					// strong: upper-left is solid
					before, after = after, before
				} else if (m.line[y][x-1] != ' ' || m.line[y-1][x] != ' ') &&
					m.line[y][x+1] != '#' && m.line[y+1][x] != '#' {
					// weak: upper-left is solid
					before, after = after, before
				}
				for dy := 0; dy < m.TileSize; dy++ {
					for dx := 0; dx < m.TileSize; dx++ {
						val := after
						if m.TileSize-dx-1 <= dy {
							val = before
						}
						floor[y*m.TileSize+dy][x*m.TileSize+dx] = val
					}
				}
			default:
				fmt.Printf("unrecognized token %c\n", t)
			case ' ':
				// "not floor" is the default setting
			}
		}
	}

	// initialize
	fromwall := make([][]uint8, 0, len(floor))
	fromfloor := make([][]uint8, 0, len(floor))
	for _, row := range floor {
		wallrow := make([]uint8, 0, len(row))
		floorrow := make([]uint8, 0, len(row))
		for _, f := range row {
			if f {
				wallrow = append(wallrow, 255)
				floorrow = append(floorrow, 0)
			} else {
				wallrow = append(wallrow, 0)
				floorrow = append(floorrow, 255)
			}
		}
		fromwall = append(fromwall, wallrow)
		fromfloor = append(fromfloor, floorrow)
	}
	// flood
	for y := 1; y < len(floor)-1; y++ {
		for x := 1; x < len(floor[y])-1; x++ {
			flood(x, y, fromwall)
			flood(x, y, fromfloor)
		}
	}

	const opaque = uint8(255)
	i := image.NewNRGBA(image.Rect(0, 0, width, height))
	m.pic = i
	offset := 0
	for y, row := range floor {
		for x, f := range row {
			if f {
				val := m.Shadow
				fw := fromwall[y][x]
				if fw > 0 {
					fw--
				}
				if int(fw) < m.WallSize/2 {
					scale := float64(m.WallSize)/2 - float64(fw)
					scale /= float64(m.WallSize)
					wall := m.WallBottom
					wall += uint8(float64(m.WallTop-m.WallBottom) * scale)
					i.Pix[offset+0] = wall
					i.Pix[offset+1] = wall
					i.Pix[offset+2] = wall
					i.Pix[offset+3] = opaque
				} else {
					fw -= uint8(m.WallSize / 2)
					dist := float64(fw)/(float64(m.TileSize)*m.ShadowWidth) + 1
					extra := uint8(float64(m.ShadowDepth) / dist)
					val += extra
					i.Pix[offset+3] = val
				}
			} else {
				ff := fromfloor[y][x]
				if ff > 0 {
					ff--
				}
				if int(ff) < (m.WallSize+1)/2 {
					scale := float64(m.WallSize+1)/2 - float64(ff)
					scale /= float64(m.WallSize)
					wall := m.WallTop
					wall -= uint8(float64(m.WallTop-m.WallBottom) * scale)
					i.Pix[offset+0] = wall
					i.Pix[offset+1] = wall
					i.Pix[offset+2] = wall
				}
				i.Pix[offset+3] = opaque
			}
			offset += 4
		}
	}
}

func flood(x int, y int, chart [][]uint8) {
	this := chart[y][x]
	if this >= 254 {
		return
	}
	this++
	if chart[y][x-1] > this {
		chart[y][x-1] = this
		if x-1 > 0 {
			flood(x-1, y, chart)
		}
	}
	if chart[y-1][x] > this {
		chart[y-1][x] = this
		if y-1 > 0 {
			flood(x, y-1, chart)
		}
	}
	if chart[y+1][x] > this {
		chart[y+1][x] = this
		if y+1 < len(chart)-1 {
			flood(x, y+1, chart)
		}
	}
	if chart[y][x+1] > this {
		chart[y][x+1] = this
		if x+1 < len(chart[y])-1 {
			flood(x+1, y, chart)
		}
	}
}
