package main

import (
	"strings"
	"testing"
)

var helloworld = []byte(strings.Join([]string{
	" package main",
	"",
	"func main() {",
	"	print(\"Hello, World!\")",
	"}",
}, string(NewLineUnix)))

func TestWindowEraseAtCursor(t *testing.T) {
	var err error
	content := helloworld
	nl_seq := []byte(NewLineUnix)
	buffer, err := bufferFromContent(content, nl_seq)
	assertNoErrors(t, err)
	window := windowFromBuffer(buffer)
	window.eraseLineAtCursor(1)
	expected := strings.Join([]string{
		"",
		"func main() {",
		"	print(\"Hello, World!\")",
		"}",
	}, string(nl_seq))

	assertBytesEqual(t, buffer.Content(), []byte(expected))
	assertBytesEqual(t, buffer.nl_seq, nl_seq)
}
