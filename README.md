# gopher-boy

[![CircleCI](https://circleci.com/gh/bokuweb/gopher-boy/tree/master.svg?style=svg)](https://circleci.com/gh/bokuweb/gopher-boy/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/bokuweb/gopher-boy)](https://goreportcard.com/report/github.com/bokuweb/gopher-boy)


<img src="screenshot/mario.png">　<img src="screenshot/tetris.png">　<img src="screenshot/drmario.png">


## Installation

you can install `gopher-boy` with following comand. 

```sh
go get github.com/bokuweb/goper-boy/cmd/gopher-boy
```

This emulator uses the go library [pixel](https://github.com/faiface/pixel), which requires OpenGL. You may need to install some requirements which can be found on the [pixels readme](https://github.com/faiface/pixel#requirements).

## Usage 

```sh
gopher-boy YOUR_GAMEBOY_ROM.gb
```

### Keymap

<kbd>&larr;</kbd> - &larr; button
<kbd>&uarr;</kbd> - &uarr; button
<kbd>&darr;</kbd> - &darr; button
<kbd>&rarr;</kbd> - &rarr; button
<kbd>Z</kbd> - A button
<kbd>X</kbd> - B button
<kbd>Enter</kbd> - Start button
<kbd>Backspace</kbd> - Select button
