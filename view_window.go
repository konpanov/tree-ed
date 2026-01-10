package main

type WindowView struct {
	window *Window
}

func (self WindowView) Draw(ctx DrawContext) {
	line_numbers_width := default_buffer_line_number_max_width(self.window.buffer)
	line_numbers_roi, main_roi := ctx.roi.SplitV(line_numbers_width)

	self.window.ResizeFrame(main_roi.Width(), main_roi.Height())

	tree_color := &TreeColorView{window: self.window}
	ln := AbsoluteLineNumberView{window: self.window}
	var cursor_view View
	switch self.window.mode {
	case InsertMode:
		cursor_view = &IndexViewCursor{window: self.window}
	case VisualMode, TreeMode:
		cursor_view = &SelectionViewCursor{window: self.window}
	default:
		cursor_view = &CharacterViewCursor{window: self.window}
	}

	main_ctx := ctx
	main_ctx.roi = main_roi
	self.DrawFrameText(main_ctx)
	tree_color.DrawNew(main_ctx)
	cursor_view.Draw(main_ctx)

	ln_ctx := ctx
	ln_ctx.roi = line_numbers_roi
	ln.DrawNew(ln_ctx)
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
