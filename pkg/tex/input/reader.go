package input

import (
	"bufio"
	"io"
)

type CircularBuffer struct {
	buffer []string
	offset int
}

func NewCircularBuffer(size int) CircularBuffer {
	return CircularBuffer{
		buffer: make([]string, size),
		offset: -1,
	}
}

func (buffer *CircularBuffer) Add(value string) {
	buffer.offset = (buffer.offset + 1) % len(buffer.buffer)
	buffer.buffer[buffer.offset] = value
}

func (buffer *CircularBuffer) Get(index int) (string, bool) {
	gap := buffer.offset - index
	if gap < 0 || gap >= len(buffer.buffer) {
		return "", false
	}
	return buffer.buffer[index%len(buffer.buffer)], true
}

type Reader struct {
	input     *bufio.Scanner
	line      []rune
	runeIndex int
	lineIndex int
	err       error
	pastLines CircularBuffer
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		input:     bufio.NewScanner(r),
		runeIndex: 1,
		lineIndex: -1,
		pastLines: NewCircularBuffer(10),
	}
}

const newlineCharacter rune = 10

func (file *Reader) ReadRune() (rune, int, error) {
	if file.err != nil {
		return 0, -1, file.err
	}
	if file.runeIndex > len(file.line) {
		if !file.input.Scan() {
			file.err = file.input.Err()
			if file.err == nil {
				file.err = io.EOF
			}
			return 0, -1, file.err
		}
		line := file.input.Text()
		file.pastLines.Add(line)
		file.line = []rune(line)
		file.lineIndex++
		file.runeIndex = 0
	}
	var result rune
	if len(file.line) == file.runeIndex || file.runeIndex == -1 {
		result = newlineCharacter
	} else {
		result = file.line[file.runeIndex]
	}
	file.runeIndex++
	return result, -1, nil
}

func (file *Reader) UnreadRune() error {
	file.runeIndex--
	return nil
}

func (file *Reader) Coordinates() (int, int) {
	return file.lineIndex, file.runeIndex-1
}

func (file *Reader) Line(index int) (string, bool) {
	return file.pastLines.Get(index)
}
