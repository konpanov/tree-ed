package main

type UndoTree struct {
	buffer        IBuffer
	modifications []UndoState
	current       int
}

type UndoState struct {
	change Change
}

func (self *UndoTree) Push(state UndoState, skip_if_empty bool) {
	if skip_if_empty && state.change.IsEmpty() {
		return
	}
	self.modifications = self.modifications[:self.current]
	self.modifications = append(self.modifications, state)
	self.current++
}

func (self *UndoTree) Curr() Change {
	if self.current == 0 {
		return nil
	}
	return self.modifications[self.current-1].change
}

func (self *UndoTree) Back() Change {
	curr := self.Curr()
	if curr != nil {
		self.current--
	}
	return curr
}

func (self *UndoTree) Forward() Change {
	if self.current == len(self.modifications) {
		return nil
	}
	self.current++
	return self.modifications[self.current-1].change
}
