package window

import (
	"github.com/bokuweb/gopher-boy/pkg/pad"
)

func NewWindow(pad *pad.Pad) *Window {
	return &Window{pad: pad}
}
