package main

import (
	"fmt"
	"image"
	"math"
	"sync"
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

type fromType struct {
	dist  uint8
	angle int16
}

func (m *MapData) draw() {
	width := 0
	height := len(m.line) * m.TileSize
	if len(m.line) > 0 {
		width = len(m.line[0]) * m.TileSize
	}

	floor := make([][]bool, height)
	angle := make([][]int16, height)
	for i := range floor {
		floor[i] = make([]bool, width)
		angle[i] = make([]int16, width)
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
				// angles
				leftempty := m.line[y][x-1] == ' '
				rightempty := m.line[y][x+1] == ' '
				aboveempty := m.line[y-1][x] == ' '
				belowempty := m.line[y+1][x] == ' '
				if leftempty {
					for dy := 0; dy < m.TileSize; dy++ {
						angle[y*m.TileSize+dy][x*m.TileSize] = 90
						angle[y*m.TileSize+dy][x*m.TileSize-1] = 90
					}
				}
				if rightempty {
					for dy := 0; dy < m.TileSize; dy++ {
						angle[y*m.TileSize+dy][(x+1)*m.TileSize-1] = 270
						angle[y*m.TileSize+dy][(x+1)*m.TileSize] = 270
					}
				}
				if aboveempty {
					for dx := 0; dx < m.TileSize; dx++ {
						angle[y*m.TileSize][x*m.TileSize+dx] = 180
						angle[y*m.TileSize-1][x*m.TileSize+dx] = 180
					}
					if !rightempty {
						angle[y*m.TileSize][(x+1)*m.TileSize-1] = 180
					}
					if !leftempty {
						angle[y*m.TileSize][x*m.TileSize] = 180
					}
				}
				if belowempty {
					for dx := 0; dx < m.TileSize; dx++ {
						angle[(y+1)*m.TileSize-1][x*m.TileSize+dx] = 360
						angle[(y+1)*m.TileSize][x*m.TileSize+dx] = 360
					}
					if !rightempty {
						angle[(y+1)*m.TileSize-1][(x+1)*m.TileSize-1] = 360
					}
					if !leftempty {
						angle[(y+1)*m.TileSize-1][x*m.TileSize] = 360
					}
				}
			case '\\':
				// lower-left is solid
				before := true
				after := false
				a := int16(180 + 45)
				if m.line[y][x+1] == '#' || m.line[y-1][x] == '#' {
					// strong rule: upper-right is solid
					before, after = after, before
					a -= 180
				} else if (m.line[y][x+1] != ' ' || m.line[y-1][x] != ' ') &&
					m.line[y][x-1] != '#' && m.line[y+1][x] != '#' {
					// weak rule: upper-right is solid if we see upper-right and don't see solid below/left
					before, after = after, before
					a -= 180
				}
				for dy := 0; dy < m.TileSize; dy++ {
					for dx := 0; dx < m.TileSize; dx++ {
						val := after
						if dx <= dy {
							val = before
						}
						floor[y*m.TileSize+dy][x*m.TileSize+dx] = val
						delta := dx - dy
						if delta <= 1 && delta >= -1 {
							angle[y*m.TileSize+dy][x*m.TileSize+dx] = a
						}
					}
				}
				if before {
					leftempty := m.line[y][x-1] == ' '
					rightempty := m.line[y][x+1] == ' '
					belowempty := m.line[y+1][x] == ' '
					if leftempty {
						for dy := 0; dy < m.TileSize; dy++ {
							angle[y*m.TileSize+dy][x*m.TileSize] = 90
							angle[y*m.TileSize+dy][x*m.TileSize-1] = 90
						}
					}
					if belowempty {
						for dx := 0; dx < m.TileSize; dx++ {
							if angle[(y+1)*m.TileSize-1][x*m.TileSize+dx] == 0 {
								angle[(y+1)*m.TileSize-1][x*m.TileSize+dx] = 360
							}
							angle[(y+1)*m.TileSize][x*m.TileSize+dx] = 360
						}
						if !rightempty {
							angle[(y+1)*m.TileSize-1][(x+1)*m.TileSize-1] = 360
						}
						if !leftempty {
							angle[(y+1)*m.TileSize-1][x*m.TileSize] = 360
						}
					}
				} else {
					leftempty := m.line[y][x-1] == ' '
					rightempty := m.line[y][x+1] == ' '
					aboveempty := m.line[y-1][x] == ' '
					if rightempty {
						for dy := 0; dy < m.TileSize; dy++ {
							if angle[y*m.TileSize+dy][(x+1)*m.TileSize-1] == 0 {
								angle[y*m.TileSize+dy][(x+1)*m.TileSize-1] = 270
							}
							angle[y*m.TileSize+dy][(x+1)*m.TileSize] = 270
						}
					}
					if aboveempty {
						for dx := 0; dx < m.TileSize; dx++ {
							if angle[y*m.TileSize][x*m.TileSize+dx] == 0 {
								angle[y*m.TileSize][x*m.TileSize+dx] = 180
							}
							angle[y*m.TileSize-1][x*m.TileSize+dx] = 180
						}
						if !rightempty {
							angle[y*m.TileSize][(x+1)*m.TileSize-1] = 180
						}
						if !leftempty {
							angle[y*m.TileSize][x*m.TileSize] = 180
						}
					}
				}
			case '/':
				// lower-right is solid
				before := true
				after := false
				a := int16(180 - 45)
				if m.line[y][x-1] == '#' || m.line[y-1][x] == '#' {
					// strong: upper-left is solid
					before, after = after, before
					a += 180
				} else if (m.line[y][x-1] != ' ' || m.line[y-1][x] != ' ') &&
					m.line[y][x+1] != '#' && m.line[y+1][x] != '#' {
					// weak: upper-left is solid
					before, after = after, before
					a += 180
				}
				for dy := 0; dy < m.TileSize; dy++ {
					for dx := 0; dx < m.TileSize; dx++ {
						val := after
						if m.TileSize-dx-1 <= dy {
							val = before
						}
						floor[y*m.TileSize+dy][x*m.TileSize+dx] = val
						delta := m.TileSize - dx - 1 - dy
						if delta <= 1 && delta >= -1 {
							angle[y*m.TileSize+dy][x*m.TileSize+dx] = a
						}
					}
				}
				if before {
					leftempty := m.line[y][x-1] == ' '
					rightempty := m.line[y][x+1] == ' '
					belowempty := m.line[y+1][x] == ' '
					if rightempty {
						for dy := 0; dy < m.TileSize; dy++ {
							if angle[y*m.TileSize+dy][(x+1)*m.TileSize-1] == 0 {
								angle[y*m.TileSize+dy][(x+1)*m.TileSize-1] = 270
							}
							angle[y*m.TileSize+dy][(x+1)*m.TileSize] = 270
						}
					}
					if belowempty {
						for dx := 0; dx < m.TileSize; dx++ {
							if angle[(y+1)*m.TileSize-1][x*m.TileSize+dx] == 0 {
								angle[(y+1)*m.TileSize-1][x*m.TileSize+dx] = 360
							}
							angle[(y+1)*m.TileSize][x*m.TileSize+dx] = 360
						}
						if !rightempty {
							angle[(y+1)*m.TileSize-1][(x+1)*m.TileSize-1] = 360
						}
						if !leftempty {
							angle[(y+1)*m.TileSize-1][x*m.TileSize] = 360
						}
					}
				} else {
					leftempty := m.line[y][x-1] == ' '
					rightempty := m.line[y][x+1] == ' '
					aboveempty := m.line[y-1][x] == ' '
					if leftempty {
						for dy := 0; dy < m.TileSize; dy++ {
							if angle[y*m.TileSize+dy][x*m.TileSize] == 0 {
								angle[y*m.TileSize+dy][x*m.TileSize] = 90
							}
							if angle[y*m.TileSize+dy][x*m.TileSize-1] == 0 {
								angle[y*m.TileSize+dy][x*m.TileSize-1] = 90
							}
						}
					}
					if aboveempty {
						for dx := 0; dx < m.TileSize; dx++ {
							if angle[y*m.TileSize][x*m.TileSize+dx] == 0 {
								angle[y*m.TileSize][x*m.TileSize+dx] = 180
							}
							angle[y*m.TileSize-1][x*m.TileSize+dx] = 180
						}
						if !rightempty {
							angle[y*m.TileSize][(x+1)*m.TileSize-1] = 180
						}
						if !leftempty {
							angle[y*m.TileSize][x*m.TileSize] = 180
						}
					}
				}
			case '>':
				before := true
				after := false
				firsta := int16(360 - 45)
				seconda := int16(180 + 45)
				if m.line[y][x+1] == '#' {
					// right is solid
					before, after = after, before
					firsta, seconda = seconda+180, firsta+180
				}
				for dy := 0; dy < m.TileSize; dy++ {
					for dx := 0; dx < m.TileSize; dx++ {
						val := after
						if dx <= dy &&
							m.TileSize-dx-1 > dy {
							val = before
						}
						floor[y*m.TileSize+dy][x*m.TileSize+dx] = val
						seta := firsta
						if dy*2 > m.TileSize {
							seta = seconda
						}
						angle[y*m.TileSize+dy][x*m.TileSize+dx] = seta
					}
				}
			case '<':
				before := true
				after := false
				firsta := int16(45)
				seconda := int16(180 - 45)
				if m.line[y][x-1] == '#' {
					// left is solid
					before, after = after, before
					firsta, seconda = seconda+180, firsta+180
				}
				for dy := 0; dy < m.TileSize; dy++ {
					for dx := 0; dx < m.TileSize; dx++ {
						val := after
						if dx > dy &&
							m.TileSize-dx-1 <= dy {
							val = before
						}
						floor[y*m.TileSize+dy][x*m.TileSize+dx] = val
						seta := firsta
						if dy*2 > m.TileSize {
							seta = seconda
						}
						angle[y*m.TileSize+dy][x*m.TileSize+dx] = seta
					}
				}
			case 'v':
				before := true
				after := false
				firsta := int16(45)
				seconda := int16(360 - 45)
				if m.line[y+1][x] == '#' {
					// below is solid
					before, after = after, before
					firsta, seconda = seconda+180, firsta+180
				}
				for dy := 0; dy < m.TileSize; dy++ {
					for dx := 0; dx < m.TileSize; dx++ {
						val := after
						if dx > dy &&
							m.TileSize-dx-1 > dy {
							val = before
						}
						floor[y*m.TileSize+dy][x*m.TileSize+dx] = val
						seta := firsta
						if dx*2 > m.TileSize {
							seta = seconda
						}
						angle[y*m.TileSize+dy][x*m.TileSize+dx] = seta
					}
				}
			case '^':
				before := true
				after := false
				firsta := int16(180 + 45)
				seconda := int16(180 - 45)
				if m.line[y-1][x] == '#' {
					// above is solid
					before, after = after, before
					firsta, seconda = seconda+180, firsta+180
				}
				for dy := 0; dy < m.TileSize; dy++ {
					for dx := 0; dx < m.TileSize; dx++ {
						val := after
						if dx <= dy &&
							m.TileSize-dx-1 <= dy {
							val = before
						}
						floor[y*m.TileSize+dy][x*m.TileSize+dx] = val
						seta := firsta
						if dx*2 > m.TileSize {
							seta = seconda
						}
						angle[y*m.TileSize+dy][x*m.TileSize+dx] = seta
					}
				}
			default:
				fmt.Printf("unrecognized token %c\n", t)
			case ' ':
				// "not floor" is the default setting
			}
		}
	}
	// fill in corners
	for y := 1; y < len(m.line)-1; y++ {
		for x := 1; x < len(m.line[0]); x++ {
			count := 0
			total := 0
			for circle := 0; circle < 4; circle++ {
				if a := angle[y*m.TileSize-(circle/2)][x*m.TileSize-(circle&1)]; a != 0 {
					count++
					total += int(a % 360)
				}
			}
			if count > 0 {
				for circle := 0; circle < 4; circle++ {
					if angle[y*m.TileSize-(circle/2)][x*m.TileSize-(circle&1)] == 0 {
						angle[y*m.TileSize-(circle/2)][x*m.TileSize-(circle&1)] = int16(total / count)
					}
				}
			}
		}
	}

	// initialize
	fromwall := make([][]fromType, 0, len(floor))
	fromfloor := make([][]fromType, 0, len(floor))
	for y, row := range floor {
		wallrow := make([]fromType, 0, len(row))
		floorrow := make([]fromType, 0, len(row))
		for x, f := range row {
			a := angle[y][x]
			wall := fromType{dist: 0, angle: a}
			floo := fromType{dist: 255, angle: a}
			if f {
				wall.dist = 255
				floo.dist = 0
			}
			wallrow = append(wallrow, wall)
			floorrow = append(floorrow, floo)
		}
		fromwall = append(fromwall, wallrow)
		fromfloor = append(fromfloor, floorrow)
	}
	// flood
	var floodwait sync.WaitGroup
	floodwait.Add(2)
	f := func(what [][]fromType) {
		for y := 3; y < len(floor)-3; y++ {
			for x := 3; x < len(floor[y])-3; x++ {
				flood(x, y, what, angle[y][x])
			}
		}
		floodwait.Done()
	}
	go f(fromwall)
	go f(fromfloor)
	floodwait.Wait()

	const opaque = uint8(255)
	const toRad = math.Pi / 180
	i := image.NewNRGBA(image.Rect(0, 0, width, height))
	m.pic = i
	offset := 0
	for y, row := range floor {
		for x, f := range row {
			if f {
				val := m.Shadow
				fw := fromwall[y][x]
				if fw.dist > 0 {
					fw.dist--
				}
				if int(fw.dist) < m.WallSize/2 {
					wallAngle := float64((m.LightAngle+360-fw.angle)%360) * toRad
					wallBoost := uint8(float64(m.Light) * math.Cos(wallAngle))
					scale := float64(m.WallSize)/2 - float64(fw.dist)
					scale /= float64(m.WallSize + 1)
					wall := m.WallBottom
					wall += uint8(float64(m.WallTop-m.WallBottom) * scale)
					wall += wallBoost
					i.Pix[offset+0] = wall
					i.Pix[offset+1] = wall
					i.Pix[offset+2] = wall
					i.Pix[offset+3] = opaque
				} else {
					fw.dist -= uint8(m.WallSize / 2)
					dist := float64(fw.dist)/(float64(m.TileSize)*m.ShadowWidth) + 1
					extra := uint8(float64(m.ShadowDepth) / dist)
					val += extra
					i.Pix[offset+3] = val
				}
			} else {
				ff := fromfloor[y][x]
				if ff.dist > 0 {
					ff.dist--
				}
				if int(ff.dist) < (m.WallSize+1)/2 {
					wallAngle := float64((m.LightAngle+360-ff.angle)%360) * toRad
					wallBoost := uint8(float64(m.Light) * math.Cos(wallAngle))
					scale := float64(m.WallSize+1)/2 - float64(ff.dist)
					scale /= float64(m.WallSize + 1)
					wall := m.WallTop
					wall -= uint8(float64(m.WallTop-m.WallBottom) * scale)
					wall += wallBoost
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

func flood(x int, y int, chart [][]fromType, setAngle int16) {
	this := chart[y][x]
	if this.dist >= 254 {
		return
	}
	this.dist++
	if chart[y][x-1].dist > this.dist {
		this.angle = setAngle
		chart[y][x-1] = this
		if x-3 > 0 {
			flood(x-1, y, chart, setAngle)
		}
	}
	if chart[y-1][x].dist > this.dist {
		this.angle = setAngle
		chart[y-1][x] = this
		if y-3 > 0 {
			flood(x, y-1, chart, setAngle)
		}
	}
	if chart[y+1][x].dist > this.dist {
		this.angle = setAngle
		chart[y+1][x] = this
		if y+3 < len(chart)-1 {
			flood(x, y+1, chart, setAngle)
		}
	}
	if chart[y][x+1].dist > this.dist {
		this.angle = setAngle
		chart[y][x+1] = this
		if x+3 < len(chart[y])-1 {
			flood(x+1, y, chart, setAngle)
		}
	}

	// Soften diagonals somewhat
	if true || this.dist >= 253 {
		return
	}
	this.dist += 2
	if chart[y-2][x-2].dist > this.dist {
		this.angle = setAngle
		chart[y-2][x-2] = this
		if x-3 > 0 && y-3 > 0 {
			flood(x-2, y-2, chart, setAngle)
		}
	}
	if chart[y-2][x+2].dist > this.dist {
		this.angle = setAngle
		chart[y-2][x+2] = this
		if x+3 < len(chart[y])-1 && y-3 > 0 {
			flood(x+2, y-2, chart, setAngle)
		}
	}
	if chart[y+2][x+2].dist > this.dist {
		this.angle = setAngle
		chart[y+2][x+2] = this
		if x+3 < len(chart[y])-1 && y+3 < len(chart)-1 {
			flood(x+2, y+2, chart, setAngle)
		}
	}
	if chart[y+2][x-2].dist > this.dist {
		this.angle = setAngle
		chart[y+2][x-2] = this
		if x-3 > 0 && y+3 < len(chart)-1 {
			flood(x-2, y+2, chart, setAngle)
		}
	}
}
