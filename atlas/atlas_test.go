package atlas

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	atl := New(512, 512)

	const iter = 1000
	var errc int
	for i := 0; i < iter; i++ {
		m := image.NewNRGBA(image.Rect(0, 0, rand.Intn(50)+10, rand.Intn(50)+10))
		r, g, b := uint8(rand.Intn(200))+55, uint8(rand.Intn(200))+55, uint8(rand.Intn(200))+55
		draw.Draw(m, m.Bounds(), image.NewUniform(color.NRGBA{R: r, G: g, B: b, A: 120}), image.ZP, draw.Src)

		if _, err := atl.Add(m); err != nil {
			errc++
		}
	}

	t.Logf("inserted %v of %v", iter-errc, iter)
	if err := atl.writeFile("out.png"); err != nil {
		t.Error(err)
	}
}

func BenchmarkAdd(b *testing.B) {
	b.ReportAllocs()
	atl := New(512, 512)
	for n := 0; n < b.N; n++ {
		src := image.NewNRGBA(image.Rect(0, 0, rand.Intn(50)+10, rand.Intn(50)+10))
		r, g, b := uint8(rand.Intn(200))+55, uint8(rand.Intn(200))+55, uint8(rand.Intn(200))+55
		draw.Draw(src, src.Bounds(), image.NewUniform(color.NRGBA{R: r, G: g, B: b, A: 120}), image.ZP, draw.Src)
		atl.Add(src)
	}
}
