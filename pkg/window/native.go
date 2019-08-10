// +build native

package window

import (
	"image/color"
	"math"

	"github.com/bokuweb/gopher-boy/pkg/constants"
	"github.com/bokuweb/gopher-boy/pkg/pad"
	"github.com/bokuweb/gopher-boy/pkg/types"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

// Window is
type Window struct {
	win   *pixelgl.Window
	image *pixel.PictureData
	pad   *pad.Pad
}


// Render renders the pixels on the window.
func (w *Window) Render(imageData types.ImageData) {
	w.image.Pix = imageData

	bg := color.RGBA{R: 0x0F, G: 0x38, B: 0x0F, A: 0xFF}
	w.win.Clear(bg)

	spr := pixel.NewSprite(pixel.Picture(w.image), pixel.R(0, 0, constants.ScreenWidth, constants.ScreenHeight))
	spr.Draw(w.win, pixel.IM)
	w.updateCamera()
	w.win.Update()
}

func (w *Window) Run(f func()) {
	pixelgl.Run(f)
}

func (w *Window) updateCamera() {
	xScale := w.win.Bounds().W() / constants.ScreenWidth
	yScale := w.win.Bounds().H() / constants.ScreenHeight
	scale := math.Min(yScale, xScale)

	shift := w.win.Bounds().Size().Scaled(0.5).Sub(pixel.ZV)
	cam := pixel.IM.Scaled(pixel.ZV, scale).Moved(shift)
	w.win.SetMatrix(cam)
}

func (w *Window) Init() {
	cfg := pixelgl.WindowConfig{
		Title:  "gopher-boy",
		Bounds: pixel.R(0, 0, constants.ScreenWidth, constants.ScreenHeight),
		// VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	win.Clear(colornames.Skyblue)
	w.win = win
	w.image = &pixel.PictureData{
		Pix:    make([]color.RGBA, constants.ScreenWidth*constants.ScreenHeight),
		Stride: constants.ScreenWidth,
		Rect:   pixel.R(0, 0, constants.ScreenWidth, constants.ScreenHeight),
	}

	// Hack: https://github.com/faiface/pixel/issues/140
	pos := win.GetPos()
	win.SetPos(pixel.ZV)
	win.SetPos(pos)
	w.updateCamera()
	win.Update()
}

func (w *Window) PollKey() {
	for key, button := range keyMap {
		if w.win.JustPressed(key) {
			w.pad.Press(button)
		}
		if w.win.JustReleased(key) {
			w.pad.Release(button)
		}
	}
}

var keyMap = map[pixelgl.Button]pad.Button{
	pixelgl.KeyZ:         pad.A,
	pixelgl.KeyX:         pad.B,
	pixelgl.KeyBackspace: pad.Select,
	pixelgl.KeyEnter:     pad.Start,
	pixelgl.KeyRight:     pad.Right,
	pixelgl.KeyLeft:      pad.Left,
	pixelgl.KeyUp:        pad.Up,
	pixelgl.KeyDown:      pad.Down,
}
