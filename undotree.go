package main

type ChangeTree struct {
	buffer        IBuffer
	modifications []Change
	current       int
}

func (self *ChangeTree) Push(mod Change) {
	if !mod.IsEmpty() {
		self.modifications = self.modifications[:self.current]
		self.modifications = append(self.modifications, mod)
		self.current++
	}
}

func (self *ChangeTree) Back() Change {
	if self.current == 0 {
		return nil
	}
	self.current--
	return self.modifications[self.current]
}

func (self *ChangeTree) Forward() Change {
	if self.current == len(self.modifications) {
		return nil
	}
	self.current++
	return self.modifications[self.current-1]
}
