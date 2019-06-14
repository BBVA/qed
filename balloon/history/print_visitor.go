/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

package history

import (
	"fmt"
	"strings"

	"github.com/bbva/qed/crypto/hashing"
)

type printVisitor struct {
	tokens []string
	height uint16
}

func newPrintVisitor(height uint16) *printVisitor {
	return &printVisitor{tokens: make([]string, 1), height: height}
}

func (v *printVisitor) Result() string {
	return fmt.Sprintf("\n%s", strings.Join(v.tokens[:], "\n"))
}

func (v *printVisitor) VisitLeafHashOp(op leafHashOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sLeafHashOp(%v)[%x]", v.indent(op.Position().Height), op.Position(), op.Value))
	return nil
}

func (v *printVisitor) VisitInnerHashOp(op innerHashOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sinnerHashOp(%v)", v.indent(op.Position().Height), op.Position()))
	op.Left.Accept(v)
	op.Right.Accept(v)
	return nil
}

func (v *printVisitor) VisitPartialInnerHashOp(op partialInnerHashOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%spartialInnerHashOp(%v)", v.indent(op.Position().Height), op.Position()))
	op.Left.Accept(v)
	return nil
}

func (v *printVisitor) VisitGetCacheOp(op getCacheOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sgetCacheOp(%v)", v.indent(op.Position().Height), op.Position()))
	return nil
}

func (v *printVisitor) VisitPutCacheOp(op putCacheOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sputCacheOp(%v)", v.indent(op.Position().Height), op.Position()))
	return op.operation.Accept(v)
}

func (v *printVisitor) VisitMutateOp(op mutateOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%smutateOp(%v)", v.indent(op.Position().Height), op.Position()))
	return op.operation.Accept(v)
}

func (v *printVisitor) VisitCollectOp(op collectOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%scollectOp(%v)", v.indent(op.Position().Height), op.Position()))
	return op.operation.Accept(v)
}

func (v printVisitor) indent(height uint16) string {
	indents := make([]string, 0)
	for i := height; i < v.height; i++ {
		indents = append(indents, "\t")
	}
	return strings.Join(indents, "")
}
