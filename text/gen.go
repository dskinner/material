// +build ignore

package main

import (
	"bytes"
	"errors"
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
	"path/filepath"
	"sort"
	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	flagOutdir   = flag.String("outdir", ".", "directory to write generated files")
	flagPkgname  = flag.String("pkgname", "text", "package name of generated go source files")
	flagFontfile = flag.String("fontfile", "", "filename of the ttf font")
	flagTSize    = flag.Int("tsize", 2048, "width and height of texture; 0 means pick smallest power of 2")
	flagFSize    = flag.Float64("fsize", 72, "font size in points")
	flagPad      = flag.Int("pad", 4, "amounted of padding for calculating sdf")
	flagScale    = flag.Int("scale", 1, "scale inputs for calculating sdf, linear resizing final ouput to inputs")
	flagBorder   = flag.Int("border", 1, "space around glyph")
	flagAscii    = flag.Bool("ascii", false, "only process ascii glyphs")
)

var ascii = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/? "

const EdgeAlpha = 0x7f

type SDF struct {
	src *image.NRGBA // where glyphs are first written
	dst *image.NRGBA // where sdf calculation is written
	out *image.NRGBA // final output, scaled if needed

	realOut *image.NRGBA

	tsize  int
	fsize  float64
	pad    int
	border int

	padr int
}

func NewSDF(textureSize int, fontSize float64, pad int, scale int, border int) *SDF {
	sdf := &SDF{
		tsize:  textureSize * scale,
		fsize:  fontSize * float64(scale),
		pad:    pad * scale,
		padr:   pad,
		border: border * scale,
	}

	sdf.src = image.NewNRGBA(image.Rect(0, 0, sdf.tsize, sdf.tsize))
	sdf.dst = image.NewNRGBA(image.Rect(0, 0, sdf.tsize, sdf.tsize))
	sdf.out = image.NewNRGBA(image.Rect(0, 0, sdf.tsize, sdf.tsize))
	draw.Draw(sdf.out, sdf.out.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)
	if scale > 1 {
		sdf.realOut = image.NewNRGBA(image.Rect(0, 0, textureSize, textureSize))
	} else {
		sdf.realOut = image.NewNRGBA(image.Rect(0, 0, sdf.tsize, sdf.tsize))
	}

	return sdf
}

func (sdf *SDF) writeSrc() {
	out, err := os.Create(filepath.Join(*flagOutdir, "src.png"))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := png.Encode(out, sdf.src); err != nil {
		log.Fatal(err)
	}
}

func (sdf *SDF) writeDst() {
	out, err := os.Create(filepath.Join(*flagOutdir, "dst.png"))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := png.Encode(out, sdf.dst); err != nil {
		log.Fatal(err)
	}
}

func (sdf *SDF) writeOut() {
	out, err := os.Create(filepath.Join(*flagOutdir, "out.png"))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// fucked up shit I'm doing just because fuck it
	// for y := 0; y < sdf.out.Bounds().Max.Y; y++ {
	// for x := 0; x < sdf.out.Bounds().Max.X; x++ {
	// c := sdf.out.At(x, y).(color.NRGBA)
	// c.R = 0xFF - c.R
	// c.G = 0xFF - c.G
	// c.B = 0xFF - c.B
	// c.A = 0x7f
	// sdf.out.Set(x, y, c)
	// }
	// }

	// TODO not sure how much of a diff this really makes... except maybe dropping the alpha channel as an end goal
	// draw.Draw(sdf.out, sdf.out.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)

	// if sdf.dst.Bounds().Eq(sdf.out.Bounds()) {
	// draw.Draw(sdf.out, sdf.out.Bounds(), sdf.dst, image.ZP, draw.Over)
	// } else {
	// b := sdf.out.Bounds()
	// TODO get back to bilinear. Originally did bilinear with alpha only image which
	// works well, but this bilinear filter with the new color images doesn't preserve
	// colors well. Could be NRGBA use, could just be the lib, but gimp does what's desired
	// with linear filter on resize.
	// rs := resize.Resize(uint(b.Dx()), uint(b.Dy()), sdf.dst, resize.NearestNeighbor)
	// draw.Draw(sdf.out, sdf.out.Bounds(), rs, image.ZP, draw.Over)
	// }

	if err := png.Encode(out, sdf.out); err != nil {
		log.Fatal(err)
	}
	sdf.writeRealOut()
}

func (sdf *SDF) writeRealOut() {
	out, err := os.Create(filepath.Join(*flagOutdir, "realout.png"))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// TODO not sure how much of a diff this really makes... except maybe dropping the alpha channel as an end goal
	draw.Draw(sdf.realOut, sdf.realOut.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)

	if sdf.dst.Bounds().Eq(sdf.realOut.Bounds()) {
		draw.Draw(sdf.realOut, sdf.realOut.Bounds(), sdf.out, image.ZP, draw.Src)
	} else {
		b := sdf.realOut.Bounds()
		// TODO get back to bilinear. Originally did bilinear with alpha only image which
		// works well, but this bilinear filter with the new color images doesn't preserve
		// colors well. Could be NRGBA use, could just be the lib, but gimp does what's desired
		// with linear filter on resize.
		rs := resize.Resize(uint(b.Dx()), uint(b.Dy()), sdf.out, resize.Bilinear)
		draw.Draw(sdf.realOut, sdf.realOut.Bounds(), rs, image.ZP, draw.Over)
	}

	if err := png.Encode(out, sdf.realOut); err != nil {
		log.Fatal(err)
	}
}

type colorSet []color.NRGBA

func (a colorSet) Len() int           { return len(a) }
func (a colorSet) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a colorSet) Less(i, j int) bool { return a[j].A < a[i].A }

func (a *colorSet) Add(c color.NRGBA) {
	for i, e := range *a {
		if c.R == e.R && c.G == e.G && c.B == e.B {
			// DO NOT TOUCH; for 2-channel color
			if (c.A <= EdgeAlpha && c.A > e.A) || (c.A > EdgeAlpha && c.A < e.A) {
				(*a)[i] = c
			}
			return
		}
	}
	*a = append(*a, c)
}

func maxu8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func minu8(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

func add(a, b uint8) uint8 {
	if x := uint32(a) + uint32(b); x <= 0xFF {
		return uint8(x)
	}
	return 0xFF
	// return a + b + 1
}

func mul(a, b uint8) uint8 {
	c := uint8((float64(a) / 0xff) * (float64(b) / 0xff) * 0xff)
	return c
}

func addclr(a, b color.NRGBA) color.NRGBA {
	a.R = mul(a.R, a.A)
	a.G = mul(a.G, a.A)
	a.B = mul(a.B, a.A)

	b.R = mul(b.R, b.A)
	b.G = mul(b.G, b.A)
	b.B = mul(b.B, b.A)

	return color.NRGBA{add(a.R, b.R), add(a.G, b.G), add(a.B, b.B), maxu8(a.A, b.A)}
}

func (a colorSet) DrawOver(img *image.NRGBA, x, y int, alpha uint8) {
	// var c color.NRGBA
	// for _, e := range a {
	// draw.Draw(img, image.Rect(x, y, x+1, y+1), &image.Uniform{color.NRGBA{R: e.R, G: e.G, B: e.B, A: alpha}}, image.ZP, draw.Over)
	// c.R += e.R
	// c.G += e.G
	// c.B += e.B
	// }
	// c.A = alpha
	// draw.Draw(img, image.Rect(x, y, x+1, y+1), &image.Uniform{c}, image.ZP, draw.Over)
}

func (a colorSet) DrawOver2(img *image.NRGBA, x, y int, alpha uint8) {
	// draw.Draw(img, image.Rect(x, y, x+1, y+1), &image.Uniform{color.Transparent}, image.ZP, draw.Src)
	// for _, e := range a {
	// draw.Draw(img, image.Rect(x, y, x+1, y+1), &image.Uniform{color.NRGBA{R: e.R, G: e.G, B: e.B, A: alpha}}, image.ZP, draw.Over)
	// }
	// at := img.At(x, y).(color.NRGBA)
	// at.A = alpha
	// draw.Draw(img, image.Rect(x, y, x+1, y+1), &image.Uniform{at}, image.ZP, draw.Src)

	c := color.NRGBA{}
	for _, e := range a {
		// e.A = alpha
		e.A = 0xFF
		c = addclr(c, e)
		c.A = alpha
	}
	// draw.Draw(img, image.Rect(x, y, x+1, y+1), &image.Uniform{color.NRGBA{B: 0xFF, A: 0xFF}}, image.ZP, draw.Src)
	// draw.Draw(img, image.Rect(x, y, x+1, y+1), &image.Uniform{c}, image.ZP, draw.Over)
	drawadd(img, x, y, c)
}

func (a colorSet) Base() color.NRGBA {
	var c color.NRGBA
	for _, e := range a {
		if e.R != 0 && e.G != 0 {
			// c.B = add(c.B, alpha)
			c.B = maxu8(c.B, mul2(0xFF, e.A))
		} else if e.G != 0 && e.B != 0 {
			// c.R = add(c.R, alpha)
			c.R = maxu8(c.R, mul2(0xFF, e.A))
		} else {
			// c.G = add(c.G, alpha)
			c.G = maxu8(c.G, mul2(0xFF, e.A))
		}
	}
	c.A = 0xFF

	return c
}

func (a colorSet) BaseCollinear() color.NRGBA {
	var c color.NRGBA
	for _, e := range a {
		if e.R != 0 && e.G != 0 {
			c.B = maxu8(c.B, 0xFF)
		} else if e.G != 0 && e.B != 0 {
			c.R = maxu8(c.R, 0xFF)
		} else {
			c.G = maxu8(c.G, 0xFF)
		}
	}
	c.A = 0xFF
	return c
}

func (a colorSet) BaseConvex(det int, dy, dx int, max int) color.NRGBA {
	// if dist > 1 {
	// log.Println(dist)
	// dist = 1
	// }

	var c color.NRGBA
	for _, e := range a {
		// if e.R != 0 && e.G != 0 {
		// c.B = maxu8(c.B, 0xFF)
		// } else if e.G != 0 && e.B != 0 {
		// c.R = maxu8(c.R, uint8(dist*0xFF))
		// } else {
		// c.G = maxu8(c.G, 0xFF)
		// }
		if e.G != 0 && e.B != 0 {
			if dx < 0 {
				if dy > max {
					log.Println(dy, max)
				}
				c.R = maxu8(c.R, uint8(float64(dy)/float64(max)*0xFF))
			}
			// if dy < 0 && dx < 0 { // Q3
			// c.R = maxu8(c.R, uint8(dist*0xFF))
			// c.G = maxu8(c.G, 0xFF-uint8(dist*0xFF))
			// } else if dy > 0 && dx < 0 { // Q2
			// c.R = maxu8(c.R, 0xFF-uint8(dist*0xFF))
			// c.G = maxu8(c.G, uint8(dist*0xFF))
			// }
		}
	}
	c.A = 0xFF
	return c
}

func mul2(a, b uint8) uint8 {
	return uint8(uint32(a) * uint32(b) / 0xFF)
}

func (a colorSet) Base3() color.NRGBA {
	var c color.NRGBA
	for _, e := range a {
		if e.R == e.G {
			c.B = maxu8(c.B, e.A)
		} else if e.B == e.R {
			c.G = maxu8(c.G, e.A)
		} else if e.G == e.B {
			c.R = maxu8(c.R, e.A)
		}
		// if e.R == e.G {
		// c.B = 0xFF
		// } else if e.B == e.R {
		// c.G = 0xFF
		// } else if e.G == e.B {
		// c.R = 0xFF
		// }

		//
		// if e.R == e.G {
		// c.B = maxu8(c.B, 0xFF-e.A)
		// } else if e.B == e.R {
		// c.G = maxu8(c.G, 0xFF-e.A)
		// } else if e.G == e.B {
		// c.R = maxu8(c.R, 0xFF-e.A)
		// }
		//
		// if e.R == e.G {
		// c.R = maxu8(c.R, e.A)
		// } else if e.B == e.R {
		// c.B = maxu8(c.B, e.A)
		// } else if e.G == e.B {
		// c.G = maxu8(c.G, e.A)
		// }
	}
	c.A = 0xFF
	return c
}

func (a colorSet) Max(alpha uint8) color.NRGBA {
	// DO NOT TOUCH; for 2-channel color
	var c color.NRGBA
	for _, e := range a {
		// TODO have access to corner color here as 1-channel value, could do something with it
		// could check orient2D of point to edge to corner to delineate...
		c.R = maxu8(c.R, mul2(e.R, e.A))
		c.G = maxu8(c.G, mul2(e.G, e.A))
		c.B = maxu8(c.B, mul2(e.B, e.A))
	}
	c.A = 0xFF
	return c
}

func (a colorSet) Max2(alpha uint8) color.NRGBA {
	// DO NOT TOUCH; for 2-channel color
	var c color.NRGBA
	// c = color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}
	for _, e := range a {
		c.R = maxu8(c.R, mul2(e.R, 0xFF-e.A))
		c.G = maxu8(c.G, mul2(e.G, 0xFF-e.A))
		c.B = maxu8(c.B, mul2(e.B, 0xFF-e.A))

		// if e.R == e.G {
		// c.R = maxu8(c.R, mul2(e.R, e.A))
		// c.G = maxu8(c.G, mul2(e.G, e.A))
		// } else if e.G == e.B {
		// c.G = maxu8(c.G, mul2(e.G, e.A))
		// c.B = maxu8(c.B, mul2(e.B, e.A))
		// } else if e.B == e.R {
		// c.B = maxu8(c.B, mul2(e.B, e.A))
		// c.R = maxu8(c.R, mul2(e.R, e.A))
		// }

		// if e.R == e.G {
		// c.R = maxu8(c.R, mul2(e.R, e.A))
		// c.G = maxu8(c.G, mul2(e.G, e.A))
		// } else if e.G == e.B {
		// c.G = maxu8(c.G, mul2(e.G, e.A))
		// c.B = maxu8(c.B, mul2(e.B, e.A))
		// } else if e.B == e.R {
		// c.B = maxu8(c.B, mul2(e.B, e.A))
		// c.R = maxu8(c.R, mul2(e.R, e.A))
		// }

		// if e.R == e.G {
		// c.R = mul2(c.R, mul2(e.R, e.A))
		// c.G = mul2(c.G, mul2(e.G, e.A))
		// } else if e.G == e.B {
		// c.G = mul2(c.G, mul2(e.G, e.A))
		// c.B = mul2(c.B, mul2(e.B, e.A))
		// } else if e.B == e.R {
		// c.B = mul2(c.B, mul2(e.B, e.A))
		// c.R = mul2(c.R, mul2(e.R, e.A))
		// }

		// if e.R == e.G {
		// c.B = maxu8(c.B, e.A)
		// c.G = maxu8(c.G, mul2(e.G, e.A))
		// c.G = 0xFF
		// } else if e.B == e.R {
		// c.G = maxu8(c.G, e.A)
		// c.R = maxu8(c.R, mul2(e.R, e.A))
		// c.R = 0xFF
		// } else if e.G == e.B {
		// c.R = maxu8(c.R, e.A)
		// c.B = maxu8(c.B, mul2(e.B, e.A))
		// c.B = 0xFF
		// }
	}

	// c = a[0]
	// c.A = alpha
	c.A = 0xFF
	return c
}

func drawadd(img *image.NRGBA, x, y int, c color.NRGBA) {
	draw.Draw(img, image.Rect(x, y, x+1, y+1), &image.Uniform{addclr(img.NRGBAAt(x, y), c)}, image.ZP, draw.Src)
}

func (sdf *SDF) calc(m image.Image) {
	var edgePoints []image.Point

	max := dist(0, 0, sdf.pad, sdf.pad) - 1
	b := m.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			ma := m.At(x, y).(color.NRGBA).A
			c := nearest(x, y, m.(*image.NRGBA).SubImage(image.Rect(x-sdf.pad, y-sdf.pad, x+sdf.pad, y+sdf.pad)))
			if c == 0xFF {
				// check if pixel is inside as a center of opposing edges
				if ma != 0 {
					sdf.dst.Set(x, y, color.RGBA{A: 0xFF})
					// sdf.dst.Set(x, y, color.White)
				}
				continue
			}

			if c == 1 && ma != 0 {
				// this point borders an edge, now determine if it's inside or out.
				edgePoints = append(edgePoints, image.Pt(x, y))
			}

			// return from nearest is always >= 1
			// decrement so that c/max returns a unit value inclusive of zero
			c--

			n := 0xFF * (1 - (float64(c) / float64(max)))
			if ma != 0 { // inside edge
				sdf.dst.Set(x, y, color.RGBA{A: 0xFF - uint8(n/2)})
			} else { // outside edge
				step := float64(0xFF) / float64(max)
				if n = n - step; n < 0 {
					n = 0
				}
				sdf.dst.Set(x, y, color.RGBA{A: uint8(n / 2)})
			}
		}
	}

	if len(edgePoints) == 0 {
		return
	}

	// for _, pt := range edgePoints {
	// alpha := sdf.dst.At(pt.X, pt.Y).(color.NRGBA).A
	// sdf.dst.Set(pt.X, pt.Y, color.RGBA{R: alpha, G: alpha, A: alpha})
	// }

	// low := edgePoints[0]
	// for _, pt := range edgePoints {
	// if pt.Y < low.Y || (pt.Y == low.Y && pt.X < low.X) {
	// low = pt
	// }
	// }

	// BUG be aware that the start point here is arbitrary and since start and end edge colors aren't verified
	// to be different, changes in drawing code or this point or otherwise might yield "bad" results even though
	// it's the direction things need to move.
	fsm := NewFSM(sdf.dst, edgePoints[0])
	// fsm.run()
	for _, pt := range edgePoints {
		c := sdf.dst.At(pt.X, pt.Y).(color.NRGBA)
		if c.R|c.G|c.B == 0 {
			fsm.dir = North
			fsm.low = pt
			fsm.run()
		}
	}

	// mark convex corners as single color; TODO should do this in fsm
	// for _, c := range fsm.corners {
	// if c.typ == Concave {
	// continue
	// }
	// clr := sdf.dst.NRGBAAt(c.pt.X, c.pt.Y)
	// if clr.R == 0 {
	// clr = color.NRGBA{R: 0xFF, A: clr.A}
	// }
	// if clr.G == 0 {
	// clr = color.NRGBA{G: 0xFF, A: clr.A}
	// }
	// if clr.B == 0 {
	// clr = color.NRGBA{B: 0xFF, A: clr.A}
	// }
	// sdf.dst.Set(c.pt.X, c.pt.Y, clr)
	// }

	// DEBUG
	// sdf.dst.Set(edgePoints[0].X, edgePoints[0].Y, color.NRGBA{0xff, 0, 0, 0xff})
	// return

	I := color.NRGBA{R: 0xFF, G: 0xFF, A: 0xFF}
	O := color.NRGBA{R: 0xFF, A: 0xFF}
	_, _ = I, O

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			curAlpha := sdf.dst.At(x, y).(color.NRGBA).A
			// find closest edge point
			min := math.MaxFloat64
			var clrSet colorSet
			var minedge image.Point
			var mincorner corner
			// a larger max distance keeps 1-channel colors from neighboring.
			maxcornerdist := math.MaxFloat64 //max //1.5 * max
			mincornerdist := maxcornerdist
			// TODO thesis seems to suggest i should be looking at closest corner, not using my closest edge approach, but who knows
			for _, edge := range edgePoints {
				// store colors for all unique edges findable in padding
				if d := dist(x, y, edge.X, edge.Y); d < max {
					c := sdf.dst.At(edge.X, edge.Y).(color.NRGBA)
					// DO NOT TOUCH; for 2-channel color
					c.A = mul2(EdgeAlpha, 0xFF-uint8(float64(d)/float64(max)*0xFF))
					clrSet.Add(c)
				}

				// find closest edge to walk down after to find closest corner along edge
				if d := dist(x, y, edge.X, edge.Y); d < min {
					// minc = sdf.dst.At(edge.X, edge.Y).(color.NRGBA)
					minedge = edge
					min = d
				}
			}

			if len(clrSet) > 0 {
				sort.Sort(clrSet)
				// minc = clrSet[0]
			}

			// find closest corner
			cornerIndices, ok := fsm.edges[minedge]
			if ok {
				i0, i1 := cornerIndices[0], cornerIndices[1]
				if i1 == len(fsm.corners) {
					i1 = 0
				}
				c0, c1 := fsm.corners[i0], fsm.corners[i1]
				d0 := dist(x, y, c0.pt.X, c0.pt.Y)
				d1 := dist(x, y, c1.pt.X, c1.pt.Y)
				if d0 < d1 {
					mincorner = c0
					mincornerdist = d0
					// minc = sdf.dst.At(c0.pt.X, c0.pt.Y).(color.NRGBA)
				} else {
					mincorner = c1
					mincornerdist = d1
					// minc = sdf.dst.At(c1.pt.X, c1.pt.Y).(color.NRGBA)
				}
			}
			_ = cornerIndices
			_ = mincornerdist
			// TODO all this needs to be cleaned up
			// for _, c := range fsm.corners {
			// if d := dist(x, y, c.pt.X, c.pt.Y); d < mincornerdist {
			// mincorner = c
			// mincornerdist = d
			// minc = sdf.dst.At(c.pt.X, c.pt.Y).(color.NRGBA)
			// }
			// }
			// _ = mincornerdist
			// tmp := NewFSM(sdf.dst, mine)
			// tmp.dir = North
			// var err error
			// LOOP:
			// for tmp.pt = tmp.low; err == nil; {
			// for _, c := range fsm.corners {
			// if tmp.pt == c.pt {
			// mincorner = c
			// mincornerdist = dist(x, y, c.pt.X, c.pt.Y)
			// minc = sdf.dst.At(c.pt.X, c.pt.Y).(color.NRGBA)
			// break LOOP
			// }
			// }
			// tmp.pt, err = tmp.next()
			// if tmp.pt.Eq(tmp.low) {
			// break
			// }
			// }

			// tmp = NewFSM(sdf.dst, mine)
			// tmp.dir = South
			// err = nil
			// LOOP2:
			// for tmp.pt = tmp.low; err == nil; {
			// for _, c := range fsm.corners {
			// if tmp.pt == c.pt {
			// d := dist(x, y, c.pt.X, c.pt.Y)
			// if d < mincornerdist {
			// mincorner = c
			// mincornerdist = d
			// minc = sdf.dst.At(c.pt.X, c.pt.Y).(color.NRGBA)
			// }
			// break LOOP2
			// }
			// }
			// tmp.pt, err = tmp.next()
			// if tmp.pt.Eq(tmp.low) {
			// break
			// }
			// }

			if min == max {
				continue
			}

			var drawClr color.NRGBA
			drawClr = clrSet.Max(curAlpha)
			drawClr.A = 0xFF

			// double up color on inside of edge so it continues increasing in intensity to mid.
			if curAlpha > EdgeAlpha {
				da := curAlpha - EdgeAlpha
				drawClr.R = add(drawClr.R, da)
				drawClr.G = add(drawClr.G, da)
				drawClr.B = add(drawClr.B, da)
				drawClr.R = add(drawClr.R, da)
				drawClr.G = add(drawClr.G, da)
				drawClr.B = add(drawClr.B, da)
			}

			if mincorner.typ == Convex && mincorner.pt == minedge {
				// if drawClr.R < drawClr.G && drawClr.R < drawClr.B {
				// drawClr.R = 0xEE //- drawClr.R
			}

			if drawClr.R != 0 {
				drawClr.R = add(drawClr.R, 9)
			}
			if drawClr.G != 0 {
				drawClr.G = add(drawClr.G, 9)
			}
			if drawClr.B != 0 {
				drawClr.B = add(drawClr.B, 9)
			}

			draw.Draw(sdf.out, image.Rect(x, y, x+1, y+1), &image.Uniform{drawClr}, image.ZP, draw.Src)
		}
	}

	// debug; draw all corners
	// for _, c := range fsm.corners {
	// typalpha := uint8(0xFF)
	// if c.typ == Collinear {
	// typalpha = 0xCC
	// }
	// if c.typ == Concave {
	// typalpha = 0x7f
	// }
	// sdf.dst.Set(c.pt.X, c.pt.Y, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: typalpha})
	// }
}

// !!!!!!!!!!!!!!!
// Let's see about finding those edges.
// Idea is everytime c == 1 add the x,y pair to a slice
// and then perform some magic kind of sort that reveals
// all the edges! Amazing!
// One possibility here is to perform the following:
//  * accumulate all points
//  * determine center of all points
//  * sort points CCW
//  * traverse slice in natural order; once direction changes, mark edge
// Determining what a "direction change" is, is another matter.
// A 90 degree or more change in direction from the starting point would
// suffice to atleast produce *something*, but that approach will have
// numerous errors. For ellipses, could watch for an axis intersection
// but that would also create a lot of frivolous edges, e.g. teardrop shape.
// Well... have to start somewhere.

// 6.1.1
// This is using the distance ratio method, so all areas, where
// another edge is less than twice as far as the closest edge, will be shared by all
// such edges. If they happen to have the same color, it will not have any effect.

// !!!!!!!!!!!!!!!

type CornerType int

const (
	Collinear CornerType = iota
	Convex
	Concave
)

type Direction image.Point

var (
	North     = Direction{0, 1}
	South     = Direction{0, -1}
	East      = Direction{1, 0}
	West      = Direction{-1, 0}
	NorthEast = Direction{1, 1}
	NorthWest = Direction{-1, 1}
	SouthEast = Direction{1, -1}
	SouthWest = Direction{-1, -1}
)

func (d Direction) Opposite() Direction {
	return Direction{-d.X, -d.Y}
}

func (d Direction) ColorIndex() int {
	return 0
	// TODO ^^^^^
	switch d {
	case North, South:
		return 0
	case East, West:
		return 1
	default:
		return 2
	}
}

type corner struct {
	pt  image.Point
	typ CornerType
	dir Direction
	det int

	left, right color.NRGBA
}

// direction from outer to inner; only useful for concave.
func (c corner) direction(m image.Image) Direction {
	for _, d := range []Direction{NorthEast, NorthWest, SouthEast, SouthWest, North, South, East, West} {
		pt := c.pt.Add(image.Point(d))
		clr := m.At(pt.X, pt.Y).(color.NRGBA)
		if clr.A < EdgeAlpha {
			return d
		}
	}
	return Direction{0, 0}
	// panic(fmt.Errorf("can't find direction for corner: %+v", c))
}

type FSM struct {
	dir Direction
	img *image.NRGBA

	low, pt image.Point
	colors  [3]color.NRGBA
	corners []corner

	edges        map[image.Point][2]int // corners index
	cornerOffset int                    // track tmp clear of corners during multiple runs for fixing edges map afterwards
}

// TODO rename EdgeDrawer and rename run to draw.
func NewFSM(img *image.NRGBA, low image.Point) *FSM {
	return &FSM{
		dir:   North,
		img:   img,
		low:   low,
		edges: make(map[image.Point][2]int),
		colors: [3]color.NRGBA{
			// color.NRGBA{R: 0x7f, G: 0x7f, A: 0x7f},
			// color.NRGBA{G: 0x7f, B: 0x7f, A: 0x7f},
			//
			// color.NRGBA{R: 0xFF, G: 0xFF, A: 0x7f},
			// color.NRGBA{G: 0xFF, B: 0xFF, A: 0x7f},
			// color.NRGBA{R: 0xFF, B: 0xFF, A: 0x7f},
			//
			color.NRGBA{G: 0xFF, B: 0xFF, A: EdgeAlpha},
			color.NRGBA{R: 0xFF, G: 0xFF, A: EdgeAlpha},
			color.NRGBA{R: 0xFF, B: 0xFF, A: EdgeAlpha},
		},
	}
}

func (fsm *FSM) run() error {
	// Draw may be called multiple times to target multiple shapes. For now, backup
	// corners from any previous call so corner types of current shape can be identified,
	// restoring before returning.
	//
	// TODO consider some type of shape struct. main (?non)issue with this right now is the
	// inner edges of a glyph like O would be considered a separate shape.
	corners := fsm.corners
	fsm.cornerOffset = len(corners)

	// TODO don't just take first point and call it a corner.
	// Should maybe do a walk without performing any other ops until first corner is found.
	// From there, start real run and pixels previously skipped will be captured with their
	// edge in its entirety.
	fsm.corners = []corner{corner{pt: fsm.low, typ: Collinear}}
	cornerOffset := fsm.cornerOffset + len(fsm.corners)
	fsm.edges[fsm.low] = [2]int{cornerOffset - 1, cornerOffset}

	var err error
	for fsm.pt = fsm.low; err == nil; {
		// if fsm.pt.Eq(fsm.low) {
		// fsm.img.Set(fsm.pt.X, fsm.pt.Y, color.NRGBA{R: 0xFF, A: 0xFF})
		// }
		fsm.pt, err = fsm.next()
		cornerOffset = fsm.cornerOffset + len(fsm.corners)
		fsm.edges[fsm.pt] = [2]int{cornerOffset - 1, cornerOffset}
		fsm.img.Set(fsm.pt.X, fsm.pt.Y, fsm.colors[fsm.dir.ColorIndex()])
		if fsm.pt.Eq(fsm.low) {
			break
		}
	}

	// determine corner types
	if n := len(fsm.corners); n > 2 {
		colors := [3]color.NRGBA{{R: 0xFF}, {G: 0xFF}, {B: 0xFF}}
		for i, c1 := range fsm.corners {
			c0, c2 := fsm.corners[bound(i-1, n)], fsm.corners[bound(i+1, n)]
			c1.det = orient2D(c0.pt, c1.pt, c2.pt)
			if c1.det < 0 {
				c1.typ = Convex
			} else if c1.det > 0 {
				c1.typ = Concave
			} else {
				// TODO delete?
			}

			//
			// TODO this is not the first run, so likely an inner portion of a glyph but not guaranteed by this check
			if len(corners) != 0 {
				c1.typ = Concave
			}
			//

			c1.dir = c1.direction(fsm.img)
			c1.left, c1.right = colors[0], colors[1]
			colors[0], colors[1], colors[2] = colors[1], colors[2], colors[0]
			fsm.corners[i] = c1
		}
	} // otherwise remain collinear

	// restore previous
	corners = append(corners, fsm.corners...)
	fsm.corners = corners

	// for k, v := range fsm.edges {
	// if v[1] == len(fsm.corners) {
	// fsm.edges[k][1] = 0 // probably not necessary
	// this is a terminating corner
	// }
	// }
	return err
}

func (fsm *FSM) at(d Direction) (image.Point, bool) {
	pt := fsm.pt.Add(image.Point(d))
	alpha := fsm.img.At(pt.X, pt.Y).(color.NRGBA).A
	return pt, alpha == (EdgeAlpha + 1)
}

// Direction
//
// Define eight directions and define each directions opposite.
//
// Choosing Color
//
// A direction and it's opposite share the same color. Looking solely at N, S, E, W,
// then that uses two of three colors required. As long as there is an intermediary
// step between a direction an it's opposite, then this remains the same.
//
// If instead a direction immediately shifts to its opposite direction, colors are
// rotated for all directions.
//
// This allows parallel lines sharing one edge to have to have the same color, in effect
// giving the line a single color. Otherwise, a sharp turn like a V shape will trigger
// a color rotation of the three available colors so that each of these edges are distinct.
//
// TODO Not much thought has been given to angled directions. Testing should be performed to
// determine what is optimal for the general case, either triggering new edges with a new
// color, or conforming to the previous direction's color to reduce unnecessary edges.
//
// from identifies the direction from where we wish to update to d.
func (fsm *FSM) setDirectionOld(d Direction, from Direction) {
	// every edge must have at least two channels on and
	// must share one of those channels with both neighboring edges.
	//
	// TODO the thesis is unclear if channel values can vary on and off or if they are strictly on and off. Seek clarification.
	//
	// BUG this doesn't address the start and end edges sharing the same color after completing the shape.
	if fsm.dir == d.Opposite() {
		fsm.colors[0], fsm.colors[1], fsm.colors[2] = fsm.colors[1], fsm.colors[2], fsm.colors[0]
	}

	// TODO determine correct corner type for use in choosing correct fade-out in image which is currently always black-to-transparent
	// but is actually determined by if this is a convex or concave.
	// may want to determine based on this, and last two, points to determine corner type. would still need something for first
	// two points, but should also consider what if this were a 1px line and the start and end points where the only marked "corners".
	// should probably default such a case to whatever looks best with median draw method and have the rest follow suit.
	//
	// BUG checking length of corners is no good as the FSM may be told to draw multiple times on separate shapes. This won't work with
	// how data is currently stored.
	//
	// TODO also need to consider updating first (and second) "corner" after completing shape, allowing type to be determined.
	//
	// corner types will be determined after draw iteration completes.
	// BUG multiple draws will cause an error given current structure. Should back up current corners and append before draw return.
	fsm.corners = append(fsm.corners, corner{pt: fsm.pt, typ: Collinear})

	// update color of current point before advance to next point during iter for consistency in defined corners.
	// choice of axis is arbitrary.
	if d == East || d == West {
		fsm.img.Set(fsm.pt.X, fsm.pt.Y, fsm.colors[d.ColorIndex()])
	}
	// ad-hoc rules to fix unexpected behavior drawing
	if (d == North && fsm.dir == East) || (d == South && fsm.dir == West) {
		fsm.img.Set(fsm.pt.X, fsm.pt.Y, fsm.colors[d.ColorIndex()])
	}

	fsm.dir = d
}

func (fsm *FSM) setDirection(d Direction, from Direction) {
	fsm.colors[0], fsm.colors[1], fsm.colors[2] = fsm.colors[1], fsm.colors[2], fsm.colors[0]
	fsm.corners = append(fsm.corners, corner{pt: fsm.pt, typ: Collinear})

	// update color of current point before advance to next point during iter for consistency in defined corners.
	// choice of axis is arbitrary.
	if d == East || d == West {
		fsm.img.Set(fsm.pt.X, fsm.pt.Y, fsm.colors[0])
	}
	// ad-hoc rules to fix unexpected behavior drawing
	if (d == North && fsm.dir == East) || (d == South && fsm.dir == West) {
		fsm.img.Set(fsm.pt.X, fsm.pt.Y, fsm.colors[0])
	}

	fsm.dir = d
}

func (fsm *FSM) corner(a image.Point) (CornerType, bool) {
	for _, c := range fsm.corners {
		if a == c.pt {
			return c.typ, true
		}
	}
	return 0, false
}

func (fsm *FSM) next() (image.Point, error) {
	switch fsm.dir {
	case North:
		if pt, ok := fsm.at(North); ok {
			return pt, nil
		}
		if pt, ok := fsm.at(NorthEast); ok {
			return pt, nil
		}
		if pt, ok := fsm.at(NorthWest); ok {
			return pt, nil
		}

		if pt, ok := fsm.at(East); ok {
			fsm.setDirection(East, East)
			return pt, nil
		}
		if pt, ok := fsm.at(West); ok {
			fsm.setDirection(West, West)
			return pt, nil
		}

		if pt, ok := fsm.at(SouthEast); ok {
			fsm.setDirection(South, SouthEast)
			return pt, nil
		}
		if pt, ok := fsm.at(South); ok {
			fsm.setDirection(South, South)
			return pt, nil
		}
		if pt, ok := fsm.at(SouthWest); ok {
			fsm.setDirection(South, SouthWest)
			return pt, nil
		}
	case East:
		if pt, ok := fsm.at(East); ok {
			return pt, nil
		}
		if pt, ok := fsm.at(SouthEast); ok {
			return pt, nil
		}
		if pt, ok := fsm.at(NorthEast); ok {
			return pt, nil
		}

		if pt, ok := fsm.at(South); ok {
			fsm.setDirection(South, South)
			return pt, nil
		}
		if pt, ok := fsm.at(North); ok {
			fsm.setDirection(North, North)
			return pt, nil
		}

		if pt, ok := fsm.at(SouthWest); ok {
			fsm.setDirection(South, SouthWest)
			return pt, nil
		}
		if pt, ok := fsm.at(NorthWest); ok {
			fsm.setDirection(North, NorthWest)
			return pt, nil
		}
		if pt, ok := fsm.at(West); ok { // check is superfluous
			fsm.setDirection(West, West)
			return pt, nil
		}
	case South:
		if pt, ok := fsm.at(South); ok {
			return pt, nil
		}
		if pt, ok := fsm.at(SouthWest); ok {
			return pt, nil
		}
		if pt, ok := fsm.at(SouthEast); ok {
			return pt, nil
		}

		if pt, ok := fsm.at(West); ok {
			fsm.setDirection(West, West)
			return pt, nil
		}
		if pt, ok := fsm.at(East); ok {
			fsm.setDirection(East, East)
			return pt, nil
		}

		if pt, ok := fsm.at(NorthWest); ok {
			fsm.setDirection(North, NorthWest)
			return pt, nil
		}
		if pt, ok := fsm.at(North); ok {
			fsm.setDirection(North, North)
			return pt, nil
		}
		// TODO maybe drop? given direction order above, this would be backtracking?
		if pt, ok := fsm.at(NorthEast); ok {
			fsm.setDirection(North, NorthEast)
			return pt, nil
		}
	case West:
		if pt, ok := fsm.at(West); ok {
			return pt, nil
		}
		if pt, ok := fsm.at(SouthWest); ok {
			return pt, nil
		}
		if pt, ok := fsm.at(NorthWest); ok {
			return pt, nil
		}

		if pt, ok := fsm.at(South); ok {
			fsm.setDirection(South, South)
			return pt, nil
		}
		if pt, ok := fsm.at(North); ok {
			fsm.setDirection(North, North)
			return pt, nil
		}

		if pt, ok := fsm.at(NorthEast); ok {
			fsm.setDirection(North, NorthEast)
			return pt, nil
		}
		if pt, ok := fsm.at(SouthEast); ok {
			fsm.setDirection(South, SouthEast)
			return pt, nil
		}
		if pt, ok := fsm.at(East); ok {
			fsm.setDirection(East, East)
			return pt, nil
		}
	}
	return image.ZP, errors.New("end of the road")
}

// nearest returns the distance to the closest pixel of
// opposite color from (mx, my) in a subspace.
func nearest(mx, my int, m image.Image) float64 {
	var min float64 = 0xFF
	ma := m.At(mx, my).(color.NRGBA).A
	b := m.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			a := m.At(x, y).(color.NRGBA).A
			// check against zero guards against bad input
			// to consistently give the desired 1px border.
			if (ma == 0 && a != 0) || (ma != 0 && a == 0) { // implicitly prevents check against itself
				dt := dist(mx, my, x, y)
				if min > dt {
					min = dt
				}
				if min == 1 { // minimum-bound reached, return early
					return min
				}
			}
		}
	}
	return min
}

// dist returns distance between two points.
func dist(x0, y0, x1, y1 int) float64 {
	x, y := x1-x0, y1-y0
	// if x < 0 {
	// x = -x
	// }
	// if y < 0 {
	// y = -y
	// }
	// if x > y {
	// return float64(x)
	// }
	// return float64(y)
	return math.Sqrt(float64(x*x + y*y))
}

// orient2D determines how points a, b, and c are arranged by the determinant
// of vectors a-c and b-c, returning a positive value if counter-clockwise,
// negative if clockwise, or zero if collinear.
func orient2D(a, b, c image.Point) int {
	m0 := a.Sub(c)
	m1 := b.Sub(c)
	return m0.X*m1.Y - m0.Y*m1.X
}

// bound wraps x to the range 0..n-1.
func bound(x int, n int) int {
	for x < 0 {
		x += n
	}
	for x >= n {
		x -= n
	}
	return x
}

type glyph struct {
	r  rune
	b  fixed.Rectangle26_6
	a  fixed.Int26_6
	tc [4]float32
}

func (g *glyph) width() int  { return (g.b.Max.X - g.b.Min.X).Ceil() }
func (g *glyph) height() int { return (g.b.Max.Y - g.b.Min.Y).Ceil() }

// area total all glyphs occupy.
func area(a []*glyph, pad int) int {
	var n int
	for _, g := range a {
		n += (g.width() + pad*2) * (g.height() + pad*2)
	}
	return n
}

// enumerate returns all glyphs with a valid index in font.
func enumerate(f *truetype.Font, fc font.Face) []*glyph {
	var gs []*glyph
	for r := rune(1); r < (1<<16)-1; r++ {
		if r == '\uFEFF' {
			continue // ignore BOM
		}
		if f.Index(r) != 0 {
			b, a, _ := fc.GlyphBounds(r)
			gs = append(gs, &glyph{r: r, b: b, a: a})
		}
	}
	return gs
}

func enumerateString(s string, fc font.Face) []*glyph {
	var gs []*glyph
	for _, r := range s {
		if b, a, ok := fc.GlyphBounds(r); ok {
			gs = append(gs, &glyph{r: r, b: b, a: a})
		}
	}
	return gs
}

// byHeight sorts glyphs tallest to shortest.
type byHeight []*glyph

func (a byHeight) Len() int           { return len(a) }
func (a byHeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byHeight) Less(i, j int) bool { return a[j].height() < a[i].height() }

func main() {
	flag.Parse()

	if *flagFontfile == "" || *flagOutdir == "" || *flagPkgname == "" {
		flag.Usage()
		return
	}

	if *flagScale < 1 {
		log.Println("scale must be >= 1")
		return
	}

	bin, err := ioutil.ReadFile(*flagFontfile)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := truetype.Parse(bin)
	if err != nil {
		log.Println(err)
		return
	}

	sdf := NewSDF(*flagTSize, *flagFSize, *flagPad, *flagScale, *flagBorder)
	d := &font.Drawer{
		Dst: sdf.src,
		Src: image.Black,
		Face: truetype.NewFace(f, &truetype.Options{
			Size:    sdf.fsize,
			Hinting: font.HintingFull,
		}),
	}

	var glyphs []*glyph
	if *flagAscii {
		glyphs = enumerateString(ascii, d.Face)
	} else {
		glyphs = enumerate(f, d.Face)
	}
	if len(glyphs) == 0 {
		log.Fatalf("sdf: failed to enumerate glyphs from %s\n", *flagFontfile)
	}
	if a := area(glyphs, sdf.pad); a > sdf.tsize*sdf.tsize {
		asq := math.Sqrt(float64(a))
		log.Fatalf("sdf: glyphs area %[1]v ~= %.2[2]fx%.2[2]f greater than texture area %[3]vx%[3]v\n", a, asq, sdf.tsize)
	}
	sort.Sort(byHeight(glyphs))

	x, y, dy := 0, 0, glyphs[0].height()+sdf.pad*2+sdf.border*2
	var wg sync.WaitGroup
	for _, g := range glyphs {
		adx, ady := g.width()+sdf.pad*2+sdf.border*2, g.height()+sdf.pad*2+sdf.border*2
		if x+adx > sdf.tsize {
			x = 0
			y += dy
			dy = ady
		}

		g.tc = [4]float32{
			// float32(x+sdf.pad+sdf.border) / float32(sdf.tsize),
			// float32(y+sdf.pad+sdf.border) / float32(sdf.tsize),
			float32(x) / float32(sdf.tsize),
			float32(y) / float32(sdf.tsize),
			float32(g.width()+sdf.pad*2+sdf.border*2) / float32(sdf.tsize),
			float32(g.height()+sdf.pad*2+sdf.border*2) / float32(sdf.tsize),
		}

		d.Dot = fixed.P(x+sdf.pad+sdf.border-int(g.b.Min.X>>6), y+sdf.pad+sdf.border-g.b.Min.Y.Ceil())
		d.DrawString(string(g.r))

		wg.Add(1)
		go func(m image.Image) {
			sdf.calc(m)
			wg.Done()
		}(sdf.src.SubImage(image.Rect(x, y, x+adx, y+ady))) // TODO workout border issues; then stop passing in dead space to sdf calc

		x += adx
	}
	wg.Wait()

	sdf.writeSrc()
	sdf.writeDst()
	sdf.writeOut()

	// generate source file to accompany out.png
	buf := new(bytes.Buffer)
	buf.WriteString("// generated by gen.go; DO NOT EDIT\n")
	fmt.Fprintf(buf, "package %s\n\n", *flagPkgname)

	ascent := float32(d.Face.Metrics().Ascent.Ceil())
	descent := float32(d.Face.Metrics().Descent.Floor())
	scale := float32(sdf.fsize) / (ascent + descent)
	fmt.Fprintf(buf, "const AscentUnit = %v\n", (ascent*scale)/float32(sdf.fsize))
	fmt.Fprintf(buf, "const DescentUnit = %v\n", (descent*scale)/float32(sdf.fsize))
	fmt.Fprintf(buf, "const TextureSize = %v\n", *flagTSize)
	fmt.Fprintf(buf, "const FontSize = %v\n", *flagFSize)
	fmt.Fprintf(buf, "const Pad = %v\n\n", *flagPad)

	buf.WriteString("var Texcoords = map[rune][4]float32{\n")
	for _, g := range glyphs {
		s := string(g.r)
		if s == "'" {
			s = `\'`
		} else if s == "\\" {
			s = `\\`
		}
		fmt.Fprintf(buf, "\t'%s': {%v, %v, %v, %v},\n", s, g.tc[0], g.tc[1], g.tc[2], g.tc[3])
	}
	buf.WriteString("}\n\n")

	buf.WriteString("var Bounds = map[rune][5]float32{\n")
	for _, g := range glyphs {
		s := string(g.r)
		if s == "'" {
			s = `\'`
		} else if s == "\\" {
			s = `\\`
		}
		nx := float32(g.b.Min.X>>6) / float32(sdf.fsize)
		ny := float32(g.b.Max.Y>>6) / float32(sdf.fsize)
		rect := g.b.Max.Sub(g.b.Min)
		w, h := float32(rect.X>>6), float32(rect.Y>>6)
		nw := float32(w) / float32(sdf.fsize)
		nh := float32(h) / float32(sdf.fsize)
		na := float32(g.a>>6) / float32(sdf.fsize)
		fmt.Fprintf(buf, "\t'%s': {%v, %v, %v, %v, %v},\n", s, nx, ny, nw, nh, na)
	}
	buf.WriteString("}\n\n")

	if err := ioutil.WriteFile(filepath.Join(*flagOutdir, "texcoords.go"), buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
