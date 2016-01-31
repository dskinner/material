// +build ignore

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "", "filename of the ttf font")
	size     = flag.Float64("size", 240, "font size in points")
	spacing  = flag.Float64("spacing", 1.2, "line spacing (e.g. 2 means double spaced)")
)

var text = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func dist(x0, y0, x1, y1 int) uint16 {
	x, y := x1-x0, y1-y0
	return uint16(math.Sqrt(float64(x*x+y*y)) + 0.5)
}

func nearest(mx, my int, m image.Image) uint16 {
	var min uint16 = 0xFFFF
	_, _, _, ma := m.At(mx, my).RGBA()

	// TODO temp fix
	if 0 < ma && ma < 0xFFFF {
		ma = 0
		m.(*image.NRGBA).Set(mx, my, color.Transparent)
	}

	b := m.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := m.At(x, y).RGBA()
			if (ma == 0 && a != 0) || (ma != 0 && a == 0) { // implicitly prevents check against itself
				dt := dist(mx, my, x, y)
				if min > dt {
					min = dt
				}
				if min == 1 { // lower-bound
					return min
				}
			}
		}
	}
	return min
}

func sdf(m image.Image) {
	var max uint16
	const ssize = 17
	size := ssize / 2
	b := m.Bounds()
	d := image.NewNRGBA64(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := nearest(x, y, m.(*image.NRGBA).SubImage(image.Rect(x-size, y-size, x+size, y+size)))
			if c != 0xFFFF {
				c--
				if max < c {
					max = c
				}
			}
			d.Set(x, y, color.RGBA64{A: c})
		}
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, ma := m.At(x, y).RGBA()
			_, _, _, da := d.At(x, y).RGBA()
			if ma == 0 && da == 0xFFFF {
				d.Set(x, y, color.RGBA64{A: 0})
				continue
			} else if ma != 0 && da == 0xFFFF {
				d.Set(x, y, color.RGBA64{A: 0xFFFF})
				continue
			}
			n := 1 - (float64(da) / float64(max))
			c := n * 0xFFFF

			if ma != 0 {
				d.Set(x, y, color.RGBA64{A: 0xFFFF - uint16(c/2)})
				// } else if ma == 0 && da == 0xFFFF {
				// d.Set(x, y, color.RGBA64{A: 0})
			} else {
				step := float64(0xFFFF) / float64(max)
				if c = c - step; c < 0 {
					c = 0
				}
				d.Set(x, y, color.RGBA64{A: uint16(c / 2)})
			}
		}
	}

	out, err := os.Create("out-sdf.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, d); err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	if *fontfile == "" {
		flag.Usage()
		return
	}

	// Read the font data.
	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	// Initialize the context.
	fg, bg := image.Black, image.Transparent
	rgba := image.NewNRGBA(image.Rect(0, 0, 2048, 2048))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	const tsize = 2048
	texcoords := make(map[rune][2]float32)

	// Draw the text.
	const inset = 20
	pt := freetype.Pt(0, int(c.PointToFixed(*size)>>6))
	for _, r := range text {
		texcoords[r] = [2]float32{
			float32(pt.X>>6) / tsize,
			float32(pt.Y>>6) / tsize,
		}
		if pt, err = c.DrawString(string(r), pt); err != nil {
			log.Println(err)
			return
		}
		if tsize-int(pt.X>>6) < int(*dpi/72**size) {
			pt.X = c.PointToFixed(0)
			pt.Y += c.PointToFixed(*size * *spacing)
		}
	}

	// signed distance func
	sdf(rgba)

	// Save that RGBA image to disk.
	outFile, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	buf := new(bytes.Buffer)
	buf.WriteString("// generated by gen.go; DO NOT EDIT\npackage text\n\n")
	buf.WriteString("var Texcoords = map[rune][2]float32{\n")
	for _, r := range text {
		tc := texcoords[r]
		fmt.Fprintf(buf, "\t'%s': {%v, %v},\n", string(r), tc[0], tc[1])
	}
	buf.WriteString("}")
	if err := ioutil.WriteFile("texcoords.go", buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
