/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

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
