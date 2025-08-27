package main

import "github.com/gdamore/tcell/v2"

type ViewEditor struct {
	editor *Editor
}

func (self *ViewEditor) Draw(ctx DrawContext) {
	status_line_height := 1
	main_roi, status_line_roi := ctx.roi.SplitH(ctx.roi.Height() - status_line_height)

	ctx.screen.Fill(' ', ctx.theme.base(tcell.StyleDefault))
	main_ctx := ctx
	main_ctx.roi = main_roi
	if self.editor.curwin == nil {
		PreviewView{}.DrawNew(main_ctx)
	} else {
		WindowView{window: self.editor.curwin}.Draw(main_ctx)
	}

	status_line_ctx := ctx
	status_line_ctx.roi = status_line_roi
	StatusLine{editor: self.editor}.DrawNew(status_line_ctx)
}
