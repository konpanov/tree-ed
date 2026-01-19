package main

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/tree-sitter/go-tree-sitter"
	sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

func TestEditorInsertNewLine(t *testing.T) {
	nl := LineBreakPosix
	buffer := mkTestBuffer(t, "ą", string(nl))
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	editor.Redraw()
	assertScreenRunes(t, editor.screen, []string{
		"1 ą       ",
		"          ",
		"[N1:1 100%",
		"          ",
	})
	OpSaveClipbaord{}.Execute(editor, 1)
	OpPasteClipboard{}.Execute(editor, 1)
	editor.Redraw()
	assertScreenRunes(t, editor.screen, []string{
		"1 ąą      ",
		"          ",
		"[N1:2 100%",
		"          ",
	})
}

func TestEditorOpenFileInWindow(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenFileInWindow("examples/twosum.c")
	editor.Redraw()
	assertScreenRunes(t, editor.screen, []string{
		"1  #includ",
		"2  #includ",
		"[N1:1   3%",
		"          ",
	})
}

func TestEditorStart(t *testing.T) {
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenFileInWindow("examples/twosum.c")
	go func() {
		editor.Start()
	}()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1  #includ",
		"2  #includ",
		"[N1:1   3%",
		"          ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1  #includ",
		"2         ",
		"[N1:1   3%",
		"          ",
	})
}

func TestEditorWordStartForward(t *testing.T) {
	content := strings.Join([]string{
		"aword bword cword",
		"dword !@# eword",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 aword bw",
		"2 dword !@",
		"[N1:1 100%",
		"          ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 aword bw",
		"2 dword !@",
		"[N1:7 100%",
		"          ",
	})
}

func TestEditorWordStartBackward(t *testing.T) {
	content := strings.Join([]string{
		"aword bword cword",
		"dword !@# eword",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 aword bw",
		"2 dword !@",
		"[N1:1 100%",
		"          ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 aword bw",
		"2 dword !@",
		"[N1:1 100%",
		"          ",
	})
}

func TestEditorWordEndForward(t *testing.T) {
	content := strings.Join([]string{
		"aword bword cword",
		"dword !@# eword",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 aword bw",
		"2 dword !@",
		"[N1:1 100%",
		"          ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 rd bword",
		"2 rd !@# e",
		"[1:11 100%",
		"          ",
	})
}

func TestEditorWordEndBackward(t *testing.T) {
	content := strings.Join([]string{
		"aword bword cword",
		"dword !@# eword",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(10, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 aword bw",
		"2 dword !@",
		"[N1:1 100%",
		"          ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'E', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 rd bword",
		"2 rd !@# e",
		"[N1:5 100%",
		"          ",
	})
}

func TestEditorLineStartText(t *testing.T) {
	content := strings.Join([]string{
		"   lalalal     ",
		"dword !@# eword",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1    lalalal        ",
		"2 dword !@# eword   ",
		"[N]         1:1 100%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, '_', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1    lalalal        ",
		"2 dword !@# eword   ",
		"[N]         1:4 100%",
		" (LF)               ",
	})
}

func TestEditorCount(t *testing.T) {
	content := strings.Join([]string{
		"   lalalal     ",
		"dword !@# eword",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1    lalalal        ",
		"2 dword !@# eword   ",
		"[N]         1:1 100%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1    lalalal        ",
		"2 dword !@# eword   ",
		"[N]        1:10 100%",
		" (LF)               ",
	})
}

func TestEditorGoToLine(t *testing.T) {
	content := strings.Join([]string{
		"line1     ",
		"line2     ",
		"line3     ",
		"line4     ",
		"line5     ",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2 line2             ",
		"[N]         1:1  40%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, '4', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"3 line3             ",
		"4 line4             ",
		"[N]         4:1  80%",
		" (LF)               ",
	})
}

func TestEditorMoveToLastLine(t *testing.T) {
	content := strings.Join([]string{
		"line1     ",
		"line2     ",
		"line3     ",
		"line4     ",
		"line5     ",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2 line2             ",
		"[N]         1:1  40%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"4 line4             ",
		"5 line5             ",
		"[N]         5:1 100%",
		" (LF)               ",
	})
}

func TestEditorSwapNodeNext(t *testing.T) {
	content := strings.Join([]string{
		"package main",
		"import (",
		" \"abc\",",
		" \"def\",",
		")",
	}, "\n")
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer := mkTestBufferWithParser(t, content, "\n", parser)
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 package main      ",
		"2 import (          ",
		"[N] ✕       1:1  40%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 main package      ",
		"2 import (          ",
		"[T] ✕       1:6  40%",
		" (LF)               ",
	})
}

func TestEditorSwapNodePrev(t *testing.T) {
	content := strings.Join([]string{
		"package main",
		"import (",
		" \"abc\",",
		" \"def\",",
		")",
	}, "\n")
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer := mkTestBufferWithParser(t, content, "\n", parser)
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 package main      ",
		"2 import (          ",
		"[N] ✕       1:1  40%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 main package      ",
		"2 import (          ",
		"[T] ✕       1:1  40%",
		" (LF)               ",
	})
}

func TestEditorDepthUp(t *testing.T) {
	content := strings.Join([]string{
		"package main",
		"import (",
		" \"abc\",",
		" \"def\",",
		")",
	}, "\n")
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer := mkTestBufferWithParser(t, content, "\n", parser)
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 4)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 package main      ",
		"2 import (          ",
		"[N] ✕       1:1  40%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"2 import            ",
		"3  \"abc\",           ",
		"[N] ✕       3:7  60%",
		" (LF)               ",
	})
}

func TestEditorDepthDown(t *testing.T) {
	content := strings.Join([]string{
		"package main",
		"import (",
		" \"abc\",",
		" \"def\",",
		")",
	}, "\n")
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer := mkTestBufferWithParser(t, content, "\n", parser)
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 8)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 package main      ",
		"2 import (          ",
		"3  \"abc\",           ",
		"4  \"def\",           ",
		"5 )                 ",
		"                    ",
		"[N] ✕       1:1 100%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 package main      ",
		"                    ",
		"                    ",
		"                    ",
		"                    ",
		"                    ",
		"[N] ✓      1:12 100%",
		" (LF)               ",
	})
}

func TestEditorEraseWordBack(t *testing.T) {
	content := strings.Join([]string{
		"package main",
		"import (",
		" \"abc\",",
		" \"def\",",
		")",
	}, "\n")
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer := mkTestBufferWithParser(t, content, "\n", parser)
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 8)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 package main      ",
		"2 import (          ",
		"3  \"abc\",           ",
		"4  \"def\",           ",
		"5 )                 ",
		"                    ",
		"[N] ✕       1:1 100%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlW, ' ', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 package           ",
		"2 import (          ",
		"3  \"abc\",           ",
		"4  \"def\",           ",
		"5 )                 ",
		"                    ",
		"[I] ✕       1:9 100%",
		" (LF)               ",
	})
}

func TestEditorEraseRuneNext(t *testing.T) {
	content := strings.Join([]string{
		"package main",
		"import (",
		" \"abc\",",
		" \"def\",",
		")",
	}, "\n")
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer := mkTestBufferWithParser(t, content, "\n", parser)
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 8)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 package main      ",
		"2 import (          ",
		"3  \"abc\",           ",
		"4  \"def\",           ",
		"5 )                 ",
		"                    ",
		"[N] ✕       1:1 100%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'I', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyDelete, ' ', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 ackage main       ",
		"2 import (          ",
		"3  \"abc\",           ",
		"4  \"def\",           ",
		"5 )                 ",
		"                    ",
		"[I] ✕       1:1 100%",
		" (LF)               ",
	})
}

func TestEditorReplaceSelection(t *testing.T) {
	content := strings.Join([]string{
		"package main",
		"import (",
		" \"abc\",",
		" \"def\",",
		")",
	}, "\n")
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(sitter_go.Language()))
	buffer := mkTestBufferWithParser(t, content, "\n", parser)
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 8)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 package main      ",
		"2 import (          ",
		"3  \"abc\",           ",
		"4  \"def\",           ",
		"5 )                 ",
		"                    ",
		"[N] ✕       1:1 100%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 hello             ",
		"2  \"abc\",           ",
		"3  \"def\",           ",
		"4 )                 ",
		"                    ",
		"                    ",
		"[I] ✕       1:6 100%",
		" (LF)               ",
	})
}

func TestEditorMoveHalfFrameDown(t *testing.T) {
	content := strings.Join([]string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"line7",
		"line8",
		"line9",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 6)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2 line2             ",
		"3 line3             ",
		"4 line4             ",
		"[N]         1:1  44%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlD, ' ', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"3 line3             ",
		"4 line4             ",
		"5 line5             ",
		"6 line6             ",
		"[N]         3:1  67%",
		" (LF)               ",
	})
}

func TestEditorMoveHalfFrameUp(t *testing.T) {
	content := strings.Join([]string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"line7",
		"line8",
		"line9",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 6)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2 line2             ",
		"3 line3             ",
		"4 line4             ",
		"[N]         1:1  44%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlD, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlD, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlU, ' ', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"3 line3             ",
		"4 line4             ",
		"5 line5             ",
		"6 line6             ",
		"[N]         3:1  67%",
		" (LF)               ",
	})
}

func TestEditorMoveFrameByLineDown(t *testing.T) {
	content := strings.Join([]string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"line7",
		"line8",
		"line9",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 6)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2 line2             ",
		"3 line3             ",
		"4 line4             ",
		"[N]         1:1  44%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"3 line3             ",
		"4 line4             ",
		"5 line5             ",
		"6 line6             ",
		"[N]         3:1  67%",
		" (LF)               ",
	})
}

func TestEditorMoveFrameByLineUp(t *testing.T) {
	content := strings.Join([]string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"line7",
		"line8",
		"line9",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 6)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2 line2             ",
		"3 line3             ",
		"4 line4             ",
		"[N]         1:1  44%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlY, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlY, ' ', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"3 line3             ",
		"4 line4             ",
		"5 line5             ",
		"6 line6             ",
		"[N]         3:1  67%",
		" (LF)               ",
	})
}

func TestEditorCenterFrame(t *testing.T) {
	content := strings.Join([]string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"line7",
		"line8",
		"line9",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 6)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2 line2             ",
		"3 line3             ",
		"4 line4             ",
		"[N]         1:1  44%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyCtrlY, ' ', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"2 line2             ",
		"3 line3             ",
		"4 line4             ",
		"5 line5             ",
		"[N]         4:1  56%",
		" (LF)               ",
	})
}

func TestEditorStartNewLineBelow(t *testing.T) {
	content := strings.Join([]string{
		"line1",
		"line2",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 6)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2 line2             ",
		"                    ",
		"                    ",
		"[N]         1:1 100%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2                   ",
		"3 line2             ",
		"                    ",
		"[I]         2:1 100%",
		" (LF)               ",
	})
}

func TestEditorStartNewLineAbove(t *testing.T) {
	content := strings.Join([]string{
		"line1",
		"line2",
	}, "\n")
	buffer := mkTestBuffer(t, content, "\n")
	screen := mkTestScreen(t, "")
	screen.SetSize(20, 6)
	editor := NewEditor(screen)
	editor.OpenBuffer(buffer)
	go func() { editor.Start() }()
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1 line1             ",
		"2 line2             ",
		"                    ",
		"                    ",
		"[N]         1:1 100%",
		" (LF)               ",
	})
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'O', tcell.ModNone))
	time.Sleep(5 * time.Millisecond)
	assertScreenRunes(t, editor.screen, []string{
		"1                   ",
		"2 line1             ",
		"3 line2             ",
		"                    ",
		"[I]         1:1 100%",
		" (LF)               ",
	})
}
