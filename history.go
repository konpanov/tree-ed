package main

type History struct {
	buffer  IBuffer
	states  []HistoryState
	current int
}

type HistoryState struct {
	change Change
}

func (self *History) Push(state HistoryState) {
	self.states = self.states[:self.current]
	self.states = append(self.states, state)
	self.current++
}

func (self *History) Curr() Change {
	if self.current == 0 {
		return nil
	}
	return self.states[self.current-1].change
}

func (self *History) Back() Change {
	curr := self.Curr()
	if curr != nil {
		self.current--
	}
	return curr
}

func (self *History) Forward() Change {
	if self.current == len(self.states) {
		return EmptyChange{}
	}
	self.current++
	return self.states[self.current-1].change
}
