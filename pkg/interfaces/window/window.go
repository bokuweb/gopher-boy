package window

// Window is
type Window interface {
	Render(imageData []byte)
	Run(run func())
	PollKey()
	KeyDown(button byte)
	KeyUp(button byte)
}
