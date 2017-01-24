// +build ignore

package main

import (
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
	if scale > 1 {
		sdf.out = image.NewNRGBA(image.Rect(0, 0, textureSize, textureSize))
	} else {
		sdf.out = image.NewNRGBA(image.Rect(0, 0, sdf.tsize, sdf.tsize))
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

	if sdf.dst.Bounds().Eq(sdf.out.Bounds()) {
		draw.Draw(sdf.out, sdf.out.Bounds(), sdf.dst, image.ZP, draw.Src)
	} else {
		b := sdf.out.Bounds()
		// TODO get back to bilinear. Originally did bilinear with alpha only image which
		// works well, but this bilinear filter with the new color images doesn't preserve
		// colors well. Could be NRGBA use, could just be the lib, but gimp does what's desired
		// with linear filter on resize.
		rs := resize.Resize(uint(b.Dx()), uint(b.Dy()), sdf.dst, resize.Bilinear)
		draw.Draw(sdf.out, sdf.out.Bounds(), rs, image.ZP, draw.Over)
	}

	if err := png.Encode(out, sdf.out); err != nil {
		log.Fatal(err)
	}
}

func (sdf *SDF) calc(m image.Image) {
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
	return math.Sqrt(float64(x*x + y*y))
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
