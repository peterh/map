package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

func read(fn string) (*MapData, error) {
	var m MapData
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	m.defaults()
	s := bufio.NewScanner(f)
	ref := reflect.Indirect(reflect.ValueOf(&m))
	for s.Scan() {
		line := s.Text()
		if strings.Contains(line, ":") {
			v := strings.SplitN(line, ":", 2)
			vref := ref.FieldByName(v[0])
			if !vref.CanSet() {
				fmt.Println("Warning:", v[0], "not a valid option")
				continue
			}
			val := strings.TrimSpace(v[1])
			switch vref.Kind() {
			case reflect.String:
				vref.SetString(val)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				i, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				vref.SetUint(i)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				vref.SetInt(i)
			case reflect.Float32, reflect.Float64:
				i, err := strconv.ParseFloat(val, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				vref.SetFloat(i)
			default:
				fmt.Println("Internal Error! Don't know how to set", val)
			}
		} else {
			m.line = append(m.line, line)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	m.rectangle()

	return &m, nil
}

func (m *MapData) rectangle() {
	// trim leading empty lines
	for len(m.line) > 0 && len(strings.TrimSpace(m.line[0])) == 0 {
		m.line = m.line[1:]
	}

	// trim trailing empty lines
	for len(m.line) > 0 && len(strings.TrimSpace(m.line[len(m.line)-1])) == 0 {
		m.line = m.line[:len(m.line)-1]
	}

	if len(m.line) == 0 {
		return
	}

	maxlen := 0
	// trim all right space (we'll square it up later)
	for i := range m.line {
		m.line[i] = strings.TrimRightFunc(m.line[i], unicode.IsSpace)
		if len(m.line[i]) > maxlen {
			maxlen = len(m.line[i])
		}
	}

	// trim unnecessary left space
	// TODO: Fix non-ASCII space (eg. NBSP is more than one byte wide, which makes this go weird)
	lefttrim := maxlen
	for _, l := range m.line {
		pos := strings.IndexFunc(l, func(r rune) bool { return !unicode.IsSpace(r) })
		if pos >= 0 {
			if pos < lefttrim {
				lefttrim = pos
			}
		}
	}
	if lefttrim > 0 {
		for i := range m.line {
			m.line[i] = m.line[i][lefttrim:]
		}
		maxlen -= lefttrim
	}

	// pad to the right
	for i := range m.line {
		if len(m.line[i]) < maxlen {
			m.line[i] = m.line[i] + strings.Repeat(" ", maxlen-len(m.line[i]))
		}
	}

	// draw border (so that 'wall' lines have somewhere to bleed over into)
	for i := range m.line {
		m.line[i] = " " + m.line[i] + " "
	}
	blank := strings.Repeat(" ", len(m.line[0]))
	m.line = append([]string{blank}, m.line...)
	m.line = append(m.line, blank)
}
