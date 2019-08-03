package bus

import "github.com/bokuweb/gopher-boy/pkg/types"

// Accessor bus accessor interface
type Accessor interface {
	WriteByte(addr types.Word, data byte)
	WriteWord(addr types.Word, data types.Word)

	ReadByte(addr types.Word) byte
	ReadWord(addr types.Word) types.Word
}
