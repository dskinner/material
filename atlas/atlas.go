package atlas

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sort"

	"dasa.cc/material/glutil"
	"dasa.cc/x/octree"
	"golang.org/x/mobile/gl"
)

var (
	DefaultFilter = glutil.TextureFilter(gl.LINEAR, gl.LINEAR)
	DefaultWrap   = glutil.TextureWrap(gl.REPEAT, gl.REPEAT)
)

type Atlas struct {
	tex     glutil.Texture
	img     *image.NRGBA
	regions []image.Rectangle
	// codes   []image.Rectangle
}

func New(w, h int) *Atlas {
	r := image.Rect(0, 0, w, h)
	return &Atlas{
		img:     image.NewNRGBA(r),
		regions: []image.Rectangle{r},
	}
}

func (atlas *Atlas) Create(ctx gl.Context) {
	atlas.tex.Create(ctx)
}

func (atlas *Atlas) Bind(ctx gl.Context) {
	atlas.tex.Bind(ctx, DefaultFilter, DefaultWrap)
}

func (atlas *Atlas) Update(ctx gl.Context) {
	if atlas.tex.Value == gl.TEXTURE0 {
		atlas.tex.Create(ctx)
		atlas.tex.Bind(ctx, DefaultFilter, DefaultWrap)
		atlas.tex.Update(ctx, 0, atlas.img.Bounds().Dx(), atlas.img.Bounds().Dy(), atlas.img.Pix)
	} else {
		atlas.tex.Bind(ctx, DefaultFilter, DefaultWrap)
		atlas.tex.Sub(ctx, 0, atlas.img.Bounds().Dx(), atlas.img.Bounds().Dy(), atlas.img.Pix)
	}
}

func (atlas *Atlas) Add(src image.Image) (image.Rectangle, error) {
	// first fit decreasing
	// TODO look at improved bin packing to replace this later on:
	// http://moose.cs.ucla.edu/publications/schreiber_korf_ijcai13.pdf
	sz := image.Rectangle{Max: src.Bounds().Size()}
	for i, r := range atlas.regions {
		s := sz.Add(r.Min)
		if s.In(r) {
			draw.Draw(atlas.img, s, src, src.Bounds().Min, draw.Over)
			// atlas.codes = append(atlas.codes, s)

			right, bottom := r, r
			right.Min.X += sz.Max.X
			bottom.Min.Y += sz.Max.Y
			if dt := r.Max.Sub(sz.Max); dt.X > dt.Y {
				bottom.Max.X = right.Min.X
			} else {
				right.Max.Y = bottom.Min.Y
			}
			atlas.regions = append(atlas.regions[:i], atlas.regions[i+1:]...)
			atlas.regions = append(atlas.regions, right, bottom)
			sort.Sort(sort.Reverse(byArea(atlas.regions)))
			return s, nil
		}
	}

	return image.ZR, fmt.Errorf("no available space to add image with size %+v", sz.Max)
}

func (atl *Atlas) writeFile(name string) error {
	out, err := os.Create(name)
	if err != nil {
		return err
	}
	defer out.Close()

	m := image.NewNRGBA(atl.img.Bounds())
	draw.Draw(m, m.Bounds(), image.NewUniform(color.White), image.ZP, draw.Src)
	draw.Draw(m, m.Bounds(), atl.img, image.ZP, draw.Over)

	return png.Encode(out, m)
}

type byArea []image.Rectangle

func (a byArea) Len() int      { return len(a) }
func (a byArea) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byArea) Less(i, j int) bool {
	p, q := a[i].Size(), a[j].Size()
	return p.X*p.Y < q.X*q.Y
}

func interleave(x0, y0, x1, y1 uint16) uint64 {
	return octree.Dilate16(y1) | octree.Dilate16(x1)<<1 | octree.Dilate16(y0)<<2 | octree.Dilate16(x0)<<3
}

func deinterleave(a uint64) (x0, y0, x1, y1 uint16) {
	return octree.Undilate16(a >> 3), octree.Undilate16(a >> 2), octree.Undilate16(a >> 1), octree.Undilate16(a)
}
