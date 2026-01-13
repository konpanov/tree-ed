package main

type WindowView struct {
	window *Window
}

func (self WindowView) Draw(ctx DrawContext) {
	line_numbers_width := default_buffer_line_number_max_width(self.window.buffer)
	line_numbers_roi, main_roi := ctx.roi.SplitV(line_numbers_width)

	self.window.ResizeFrame(main_roi.Width(), main_roi.Height())

	main_ctx := ctx
	main_ctx.roi = main_roi

	self.DrawFrameText(main_ctx)

	tree_color := &TreeView{window: self.window}
	tree_color.Draw(main_ctx)

	var cursor_view View
	switch self.window.mode {
	case InsertMode:
		cursor_view = &EdgeCursorView{window: self.window}
		cursor_view.Draw(main_ctx)
	case VisualMode, TreeMode:
		cursor_view = &RangeView{window: self.window}
		cursor_view.Draw(main_ctx)
		cursor_view = &CharCursorView{window: self.window}
		cursor_view.Draw(main_ctx)
	default:
		cursor_view = &CharCursorView{window: self.window}
		cursor_view.Draw(main_ctx)
	}

	ln := LineNumberView{window: self.window}
	ln_ctx := ctx
	ln_ctx.roi = line_numbers_roi
	ln.Draw(ln_ctx)

}

func (self WindowView) DrawFrameText(ctx DrawContext) {
	frame := self.window.frame
	offset := frame.TopLeft()
	cursor := self.window.cursor.AsEdge().MoveToRunePos(offset)
	for !cursor.IsEnd() {
		pos := cursor.Pos()
		rel_pos := frame.RelativePosition(pos)
		if rel_pos == Below {
			break
		}
		if rel_pos == Inside {
			r, _ := cursor.Rune()
			for _, value := range RenderedRune(r) {
				screen_pos := text_pos_to_screen(pos, offset, ctx.roi)
				set_rune(ctx.screen, screen_pos, value)
			}
		}
		cursor = cursor.RuneNext()
	}
}
