package utils

import (
	"bufio"
	"os"
)

// LoadROM loads gameboy ROMfile to buf
func LoadROM(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}
	size := stats.Size()
	bytes := make([]byte, size)
	b := bufio.NewReader(file)
	_, err = b.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
