package utils

import "github.com/bokuweb/gopher-boy/pkg/types"

func Bytes2Word(upper, lower byte) types.Word {
	return (types.Word(upper) << 8) ^ types.Word(lower)
}

func Word2Bytes(w types.Word) (byte, byte) {
	return byte(w >> 8), byte(w & 0x00FF)
}
