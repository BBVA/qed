package pruning2

type OperationsStack []*Operation

func NewOperationsStack() *OperationsStack {
	return new(OperationsStack)
}

func (s *OperationsStack) Len() int {
	return len(*s)
}

func (s OperationsStack) Peek() (op *Operation) {
	return s[len(s)-1]
}

func (s *OperationsStack) Pop() (op *Operation) {
	i := s.Len() - 1
	op = (*s)[i]
	*s = (*s)[:i]
	return
}

func (s *OperationsStack) Push(op *Operation) {
	*s = append(*s, op)
}

func (s *OperationsStack) PushAll(ops ...*Operation) {
	*s = append(*s, ops...)
}

func (s *OperationsStack) List() []*Operation {
	l := make([]*Operation, 0)
	for s.Len() > 0 {
		l = append(l, s.Pop())
	}
	return l
}
