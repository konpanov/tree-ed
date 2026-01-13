package main

import "github.com/gdamore/tcell/v2"

type EditorView struct {
	editor *Editor
}

func (self *EditorView) Draw(ctx DrawContext) {
	status_line_height := 2
	main_roi, status_line_roi := ctx.roi.SplitH(ctx.roi.Height() - status_line_height)

	ctx.screen.Fill(' ', ctx.theme.base(tcell.StyleDefault))
	main_ctx := ctx
	main_ctx.roi = main_roi
	if self.editor.curwin == nil {
		PreviewView{}.Draw(main_ctx)
	} else {
		WindowView{window: self.editor.curwin}.Draw(main_ctx)
	}

	status_line_ctx := ctx
	status_line_ctx.roi = status_line_roi
	StatusLineView{editor: self.editor}.Draw(status_line_ctx)
}
