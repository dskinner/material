package material

import (
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

type Behavior int

const (
	// When screen space is available, a surface is always visible.
	VisibilityPermanent Behavior = iota

	// Surface visibility can be toggled between visible and hidden. When visible,
	// interacting with other elements on the screen does not change visibility.
	VisibilityPersistent

	// Surface visibility can be toggled between visible and hidden. When visible,
	// interacting with other elements on the screen toggles the surface to become
	// hidden or minimized.
	VisibilityTemporary

	// Element width stays the same when screen size changes.
	WidthFixed

	// Element width grows as screen size changes.
	WidthFluid

	// Element width is fixed until influenced by another element or breakpoint.
	WidthSticky

	// Element width contracts as a panel is revealed
	WidthSqueeze

	// Element width stays the same, its position changes horizontally as a panel
	// appears, and it may be partially occluded by a screen’s edge.
	WidthPush

	// Element width and position stays the same as a panel appears over content.
	WidthOverlay

	// The z position, and shadow of an element. A flat element will have no shadow.
	DescriptorFlat
	DescriptorRaised
)

type Grid struct {
	Margin  float32
	Gutter  float32
	Columns int

	debug *Material
}

func (gd *Grid) StepSize() float32 {
	return (float32(windowSize.WidthPx) - (gd.Margin * 2)) / float32(gd.Columns)
}

// TODO avoid the pointer
func NewGrid() *Grid {
	// by breakpoints
	g := &Grid{Margin: 24, Gutter: 24, Columns: 12}
	if windowSize.WidthPx < 600 || windowSize.HeightPx < 600 {
		g.Margin, g.Gutter = 16, 16 // TODO dp vals
	}
	if windowSize.WidthPx < 480 {
		g.Columns = 4
	} else if windowSize.WidthPx < 720 {
		g.Columns = 8
	}
	return g
}

func (gd *Grid) draw(ctx gl.Context, view, proj f32.Mat4) {
	if gd.debug == nil {
		gd.debug = New(ctx, Color(0x03A9F499))
		gd.debug.world.Identity()
		gd.debug.world[0][0] = gd.Gutter + gd.Margin
		gd.debug.world[1][1] = float32(windowSize.HeightPx)
	}

	step := gd.StepSize()
	for i := 0; i <= gd.Columns; i++ {
		if i == 0 {
			gd.debug.world[0][0] = gd.Margin
			gd.debug.world[0][3] = 0
		} else if i == gd.Columns {
			gd.debug.world[0][0] = gd.Margin
			gd.debug.world[0][3] = gd.Margin + float32(i)*step
		} else {
			gd.debug.world[0][0] = gd.Gutter
			gd.debug.world[0][3] = gd.Margin + float32(i)*step - gd.Gutter/2.0
		}
		gd.debug.Draw(ctx, view, proj)
	}
}
