package main

import (
	"slices"
	"strings"
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

func TestOperationNone(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenFileInWindow("examples/twosum.c")
	windowBefore := *editor.curwin
	OpNone{}.Execute(editor, 1)
	windowAfter := *editor.curwin
	if windowBefore != windowAfter {
		t.Error("OpNone should not change window state")
	}
}

func TestOperationCursorDown(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("hello\nbye\n"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	OpCursorDown{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 6; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(4), true)
	OpCursorDown{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 8; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOperationCursorUp(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\n"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(4), true)
	OpCursorUp{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 0; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(8), true)
	OpCursorUp{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 2; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOperationCursorLeft(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\nwhat"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(6), true)
	OpCursorLeft{}.Execute(editor, 4)
	if index, expected := editor.curwin.cursor.Index(), 4; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOperationCursorRight(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\nwhat"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(6), true)
	OpCursorRight{}.Execute(editor, 4)
	if index, expected := editor.curwin.cursor.Index(), 8; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOperationInsertBeforeCursor(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\nwhat"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(6), true)
	OpInsertBeforeCursor{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 6; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if !editor.curwin.cursor.as_edge {
		t.Error("Expected window cursor to be edge")
	}
}

func TestOperationInsertAfterCursor(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\nwhat"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(8), true)
	OpInsertAfterCursor{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 9; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if !editor.curwin.cursor.as_edge {
		t.Error("Expected window cursor to be edge")
	}
}

func TestOperationInsertAfterLine(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\nwhat"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(5), true)
	OpInsertAfterLine{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 9; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if !editor.curwin.cursor.as_edge {
		t.Error("Expected window cursor to be edge")
	}
}

func TestOperationInsertBeforeLine(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\nwhat"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(7), true)
	OpInsertBeforeLine{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 4; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if !editor.curwin.cursor.as_edge {
		t.Error("Expected window cursor to be edge")
	}
}

func TestOperationVisual(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\nwhat"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(7), true)
	OpVisual{}.Execute(editor, 1)
	OpCursorUp{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 2; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := editor.curwin.anchor.Index(), 7; index != expected {
		t.Errorf("Unexpected anchor index %d, expected %d", index, expected)
	}
	if editor.curwin.cursor.as_edge {
		t.Error("Expected window cursor to be char")
	}
}

func TestOprationVisualAsAnchor(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\nwhat"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(4), true)
	editor.curwin.setAnchor(editor.curwin.anchor.ToIndex(8))
	OpVisualAsAnchor{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 8; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := editor.curwin.anchor.Index(), 4; index != expected {
		t.Errorf("Unexpected anchor index %d, expected %d", index, expected)
	}
	if editor.curwin.cursor.as_edge {
		t.Error("Expected window cursor to be char")
	}
}

func TestOprationNormalAsAnchor(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	buffer, err := bufferFromContent([]byte("bye\nhello\nwhat"), []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	OpVisual{}.Execute(editor, 1)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(4), true)
	editor.curwin.setAnchor(editor.curwin.anchor.ToIndex(8))
	OpNormalAsAnchor{}.Execute(editor, 1)
	if index, expected := editor.curwin.cursor.Index(), 8; index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := editor.curwin.anchor.Index(), 8; index != expected {
		t.Errorf("Unexpected anchor index %d, expected %d", index, expected)
	}
	if editor.curwin.cursor.as_edge {
		t.Error("Expected window cursor to be char")
	}
}

func TestOprationTree(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	OpTree{}.Execute(editor, 1)
	start, end := editor.curwin.getSelection()
	if index, expected := start, uint(0); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := end, uint(7); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOprationCopyCursorLine(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	buffer, err := bufferFromContent(content, []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	OpCopyCursorLine{}.Execute(editor, 1)
	OpPasteClipboard{}.Execute(editor, 1)
	expected := []byte(strings.Join([]string{
		"package main",
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	recevied := editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
}

func TestOpeartionEraseRune(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	buffer, err := bufferFromContent(content, []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	OpEraseRune{}.Execute(editor, 1)
	expected := []byte(strings.Join([]string{
		"ackage main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	recevied := editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
	OpUndoChange{}.Execute(editor, 1)
	expected = []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	recevied = editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
	OpRedoChange{}.Execute(editor, 1)
	expected = []byte(strings.Join([]string{
		"ackage main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	recevied = editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
}

func TestOpeartionEraseRunePrev(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	buffer, err := bufferFromContent(content, []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(6), true)
	OpInsertBeforeCursor{}.Execute(editor, 1)
	OpEraseRunePrev{}.Execute(editor, 1)
	expected := []byte(strings.Join([]string{
		"packae main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	recevied := editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
	if pos, expected := editor.curwin.cursor.Index(), 5; pos != expected {
		t.Errorf("Unexpected cursor position. Expected %d, got %d.\n", expected, pos)
	}
	OpUndoChange{}.Execute(editor, 1)
	expected = []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	recevied = editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
	if pos, expected := editor.curwin.cursor.Index(), 5; pos != expected {
		t.Errorf("Unexpected cursor position. Expected %d, got %d.\n", expected, pos)
	}
	OpRedoChange{}.Execute(editor, 1)
	expected = []byte(strings.Join([]string{
		"packae main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	recevied = editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
	if pos, expected := editor.curwin.cursor.Index(), 5; pos != expected {
		t.Errorf("Unexpected cursor position. Expected %d, got %d.\n", expected, pos)
	}
}

func TestOpeartionInsertInput(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	buffer, err := bufferFromContent(content, []byte("\n"), nil)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(7), true)
	OpInsertBeforeCursor{}.Execute(editor, 1)
	OpInsertInput{lines: [][]byte{[]byte(" 123")}}.Execute(editor, 1)
	expected := []byte(strings.Join([]string{
		"package 123 main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	recevied := editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
	if pos, expected := editor.curwin.cursor.Index(), 11; pos != expected {
		t.Errorf("Unexpected cursor position. Expected %d, got %d.\n", expected, pos)
	}
	OpInsertInput{lines: [][]byte{
		[]byte(" 456"),
		[]byte(" 789"),
		[]byte(" abc"),
	}}.Execute(editor, 1)
	expected = []byte(strings.Join([]string{
		"package 123 456",
		" 789",
		" abc main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	recevied = editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
	if pos, expected := editor.curwin.cursor.Index(), 25; pos != expected {
		t.Errorf("Unexpected cursor position. Expected %d, got %d.\n", expected, pos)
	}
}

func TestOprationNodeUp(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	OpTree{}.Execute(editor, 1)
	OpNodeUp{}.Execute(editor, 1)
	start, end := editor.curwin.getSelection()
	if index, expected := start, uint(0); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := end, uint(12); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOprationNodeNextSibling(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	OpTree{}.Execute(editor, 1)
	OpNodeNextSibling{}.Execute(editor, 1)
	start, end := editor.curwin.getSelection()
	if index, expected := start, uint(8); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := end, uint(12); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOprationNodeDown(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	OpTree{}.Execute(editor, 1)
	OpNodeUp{}.Execute(editor, 1)
	OpNodeNextSibling{}.Execute(editor, 1)
	OpNodeDown{}.Execute(editor, 1)
	start, end := editor.curwin.getSelection()
	if index, expected := start, uint(13); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := end, uint(19); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOprationNodePrevSibling(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(9), true)
	OpTree{}.Execute(editor, 1)
	OpNodePrevSibling{}.Execute(editor, 1)
	start, end := editor.curwin.getSelection()
	if index, expected := start, uint(0); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := end, uint(7); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOprationNodeNextDepth(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(9), true)
	OpTree{}.Execute(editor, 1)
	OpNodeNextDepth{}.Execute(editor, 1)
	start, end := editor.curwin.getSelection()
	if index, expected := start, uint(13); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := end, uint(19); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOprationNodePrevDepth(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(18), true)
	OpTree{}.Execute(editor, 1)
	OpNodePrevDepth{}.Execute(editor, 1)
	start, end := editor.curwin.getSelection()
	if index, expected := start, uint(8); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := end, uint(12); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOprationNodeFirstSibling(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		"    \"abc\"",
		"    \"efg\"",
		"    \"hij\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(57), true)
	OpTree{}.Execute(editor, 1)
	OpNodeUp{}.Execute(editor, 1)
	OpNodeUp{}.Execute(editor, 1)
	OpNodeFirstSibling{}.Execute(editor, 1)
	start, end := editor.curwin.getSelection()
	if index, expected := start, uint(20); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := end, uint(21); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOprationNodeLastSibling(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		"    \"abc\"",
		"    \"efg\"",
		"    \"hij\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(57), true)
	OpTree{}.Execute(editor, 1)
	OpNodeUp{}.Execute(editor, 1)
	OpNodeUp{}.Execute(editor, 1)
	OpNodeLastSibling{}.Execute(editor, 1)
	start, end := editor.curwin.getSelection()
	if index, expected := start, uint(80); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
	if index, expected := end, uint(81); index != expected {
		t.Errorf("Unexpected index %d, expected %d", index, expected)
	}
}

func TestOprationNodeEraseSelection(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	content := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		"    \"abc\"",
		"    \"efg\"",
		"    \"hij\"",
		")",
	}, "\n"))
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer, err := bufferFromContent(content, []byte("\n"), parser)
	assertNoErrors(t, err)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.curwin.setCursor(editor.curwin.cursor.ToIndex(57), true)
	OpTree{}.Execute(editor, 1)
	OpNodeUp{}.Execute(editor, 1)
	OpNodeUp{}.Execute(editor, 1)
	OpEraseSelection{}.Execute(editor, 1)

	expected := []byte(strings.Join([]string{
		"package main",
		"import (",
		"    \"strings\"",
		"    \"testing\"",
		"    ",
		"    \"efg\"",
		"    \"hij\"",
		")",
	}, "\n"))
	recevied := editor.curwin.buffer.Content()
	if slices.Compare(expected, recevied) != 0 {
		t.Errorf("Unexpected buffer content. Expected: \"%s\" \n Recieved: \"%s\"\n", expected, recevied)
	}
	if pos, expected := editor.curwin.cursor.Index(), 53; pos != expected {
		t.Errorf("Unexpected cursor position. Expected %d, got %d.\n", expected, pos)
	}
}
