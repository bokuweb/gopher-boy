package window

import "github.com/bokuweb/gopher-boy/pkg/types"

// Window is
type Window interface {
	Render(imageData types.ImageData)
	Run(run func())
	PollKey()
}
