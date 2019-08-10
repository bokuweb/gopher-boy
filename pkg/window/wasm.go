// +build wasm

package window

import (
	"github.com/bokuweb/gopher-boy/pkg/pad"
	"github.com/bokuweb/gopher-boy/pkg/types"
)

// Window is
type Window struct {
	pad   *pad.Pad
}

// Render renders the pixels on the window.
func (w *Window) Render(imageData types.ImageData) {
}

func (w *Window) Run(f func()) {
}

func (w *Window) Init() {
}

func (w *Window) PollKey() {
	// for key, button := range keyMap {
	// 	if w.win.JustPressed(key) {
	// 		w.pad.Press(button)
	// 	}
	// 	if w.win.JustReleased(key) {
	// 		w.pad.Release(button)
	// 	}
	// }
}

// var keyMap = map[pixelgl.Button]pad.Button{
// 	pixelgl.KeyZ:         pad.A,
// 	pixelgl.KeyX:         pad.B,
// 	pixelgl.KeyBackspace: pad.Select,
// 	pixelgl.KeyEnter:     pad.Start,
// 	pixelgl.KeyRight:     pad.Right,
// 	pixelgl.KeyLeft:      pad.Left,
// 	pixelgl.KeyUp:        pad.Up,
// 	pixelgl.KeyDown:      pad.Down,
// }
