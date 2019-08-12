// +build wasm

package window

import (
	"github.com/bokuweb/gopher-boy/pkg/pad"
	"github.com/bokuweb/gopher-boy/pkg/types"
)

// Window is
type Window struct {
	keyState byte
	pad      *pad.Pad
}

// Render renders the pixels on the window.
func (w *Window) Render(imageData types.ImageData) {
	/* NOP */
}

func (w *Window) Run(f func()) {
	/* NOP */
}

func (w *Window) Init() {
	/* NOP */
}

func (w *Window) PollKey() {
	i := byte(0)
	for i < 8 {
		b := byte(0x01 << i)
		if w.keyState&b != 0 {
			w.pad.Press(pad.Button(b))
		} else {
			w.pad.Release(pad.Button(b))
		}
		i++
	}
}

func (w *Window) KeyDown(button byte) {
	w.keyState |= byte(button)
}

func (w *Window) KeyUp(button byte) {
	w.keyState &= ^byte(button)
}
