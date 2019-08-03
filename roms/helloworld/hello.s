	INCLUDE "hardware.inc"

	SECTION	"Start",ROM0[$100]		; start vector, followed by header data applied by rgbfix.exe
	nop
	jp	Start

        SECTION "Example",ROM0[$150]		; code starts here

Start:
	di					; disable interrupts
	ld	sp,$E000			; setup stack

.wait_vbl
	ld	a,[rLY]	
	cp	$90
	jr	nz,.wait_vbl			; wait for vblank to properly disable lcd

	xor	a
	ld	[rIF],a				; reset usual regs
	ld	[rLCDC],a
	ld	[rSTAT],a
	ld	[rSCX],a
	ld	[rSCY],a
	ld	[rLYC],a
	ld	[rIE],a

	ld	hl,_RAM                         ; fill ram with a, a = 0 here
	ld	bc,$2000-2			; watch out for stack ;)
	call	fill

	ld	hl,_HRAM
	ld	c,$80				; a = 0, b = 0 here, so let's save a cycle or two
	call	fill

	ld	hl,_VRAM
	ld	b,$18				; a = 0, bc should be $1800; c = 0 here, so..
	call	fill

	ld	a,$20				; ascii code for 'space'
	; no need to setup hl since _SCRN0 and _SCRN1 are part of _VRAM, just continue
	ld	b,8				; bc should be $800; c = 0 here, so..
	call	fill

	ld	a,%10010011			; 00 - light, 01 - gray, 10 - dark grey, 11 - dark
	ld	[rBGP],a			; bg palette
	ld	[rOBP0],a			; obj palettes
	ld	[rOBP1],a

	ld	hl,font				; font data
	ld	de,_VRAM+$200			; place it here to get ascii mapping
	ld	bc,1776				; speccy.chr file size
	call 	copy

	ld	hl,text				; hello message
	ld	de,_SCRN0+$100			; center it a bit
	ld	c,text_end-text			; b = 0, our string = 18 chars, so..
	call	copy				; lcdc is disabled so you have 'easy' access to vram
	
	ld	a,LCDCF_ON | LCDCF_BG8000 | LCDCF_BG9800 | LCDCF_OBJ8 | LCDCF_OBJOFF | LCDCF_WINOFF | LCDCF_BGON
	ld	[rLCDC],a			; enable lcd

.the_end
	halt
	nop

	jr	.the_end

;-------------------------------------------------------------------------------	
copy:
;-------------------------------------------------------------------------------
; hl - source address
; de - destination
; bc - size
	inc	b
	inc	c
	jr	.skip
.copy
	ld	a,[hl+]
	ld	[de],a
	inc	de
.skip
	dec	c
	jr	nz,.copy
	dec	b
	jr	nz,.copy
	ret

;-------------------------------------------------------------------------------
fill:
;-------------------------------------------------------------------------------
; a - byte to fill with
; hl - destination address
; bc - size of area to fill

	inc	b
	inc	c
	jr	.skip
.fill
	ld	[hl+],a
.skip
	dec	c
	jr	nz,.fill
	dec	b
	jr	nz,.fill
	ret

;-------------------------------------------------------------------------------

font:
        INCBIN	"speccy.chr"					; created with https://github.com/gitendo/bmp2cgb

text:
	DB	" Hello 8-bit world! "
text_end: