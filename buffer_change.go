package main

import (
	"slices"

	sitter "github.com/smacker/go-tree-sitter"
)

type BufferChange struct {
	start_index   int
	new_end_index int
	old_end_index int
	start_pos     Point
	new_end_pos   Point
	old_end_pos   Point

	before []byte
	after  []byte
}

func (self BufferChange) Reverse() BufferChange {
	return BufferChange{
		before:        self.after,
		after:         self.before,
		start_index:   self.start_index,
		new_end_index: self.old_end_index,
		old_end_index: self.new_end_index,
		start_pos:     self.start_pos,
		new_end_pos:   self.old_end_pos,
		old_end_pos:   self.new_end_pos,
	}
}

func (self BufferChange) Undo(content []byte) []byte {
	return slices.Replace(content, self.start_index, self.new_end_index, self.before...)
}

func (self BufferChange) Equal(other BufferChange) bool {
	return slices.Compare(self.before, other.before) == 0 &&
		slices.Compare(self.after, other.after) == 0 &&
		self.start_index == other.start_index &&
		self.new_end_index == other.new_end_index &&
		self.old_end_index == other.old_end_index &&
		self.start_pos == other.start_pos &&
		self.new_end_pos == other.new_end_pos &&
		self.old_end_pos == other.old_end_pos
}

func (first BufferChange) Merge(second BufferChange) (BufferChange, error) {
	if len(first.before) == 0 && len(second.before) == 0 { // Merge inserts
		if first.new_end_index != second.start_index {
			return first, ErrChangesAreNotContinuous
		}
		return BufferChange{
			before:        append(first.before, second.before...),
			after:         append(first.after, second.after...),
			start_index:   first.start_index,
			old_end_index: first.old_end_index,
			new_end_index: second.new_end_index,
			start_pos:     first.start_pos,
			old_end_pos:   first.old_end_pos,
			new_end_pos:   second.new_end_pos,
		}, nil
	} else if len(first.after) == 0 && len(second.after) == 0 { // Merge erases
		if first.start_index != second.old_end_index {
			return first, ErrChangesAreNotContinuous
		}
		// testabcde
		//       ^^
		// testabde
		//      ^^
		// testade
		return BufferChange{
			before:        append(second.before, first.before...),
			after:         append(first.after, second.after...),
			start_index:   second.start_index,
			old_end_index: first.old_end_index,
			new_end_index: second.new_end_index,
			start_pos:     second.start_pos,
			old_end_pos:   first.old_end_pos,
			new_end_pos:   second.new_end_pos,
		}, nil
	}
	return first, ErrChangesAreNotContinuous
}

func (sefl BufferChange) ToSitterEditInput() sitter.EditInput {
	edit := sitter.EditInput{}
	edit.StartIndex = uint32(sefl.start_index)
	edit.OldEndIndex = uint32(sefl.old_end_index)
	edit.NewEndIndex = uint32(sefl.new_end_index)
	edit.StartPoint = sitterPoint(sefl.start_pos)
	edit.OldEndPoint = sitterPoint(sefl.old_end_pos)
	edit.NewEndPoint = sitterPoint(sefl.new_end_pos)
	return edit
}
