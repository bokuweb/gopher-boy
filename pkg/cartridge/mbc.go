package cartridge

import (
	"github.com/bokuweb/gopher-boy/pkg/types"
)

type MBC interface {
	Write(addr types.Word, value byte)
	Read(addr types.Word) byte
	switchROMBank(bank int)
	switchRAMBank(bank int)
}
