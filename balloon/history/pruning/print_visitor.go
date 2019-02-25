package pruning

import (
	"fmt"
	"strings"

	"github.com/bbva/qed/hashing"
)

type PrintVisitor struct {
	tokens []string
	height uint16
}

func NewPrintVisitor(height uint16) *PrintVisitor {
	return &PrintVisitor{tokens: make([]string, 1), height: height}
}

func (v *PrintVisitor) Result() string {
	return fmt.Sprintf("\n%s", strings.Join(v.tokens[:], "\n"))
}

func (v *PrintVisitor) VisitLeafHashOp(op LeafHashOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sLeafHashOp(%v)[%x]", v.indent(op.Position().Height), op.Position(), op.Value))
	return nil
}

func (v *PrintVisitor) VisitInnerHashOp(op InnerHashOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sInnerHashOp(%v)", v.indent(op.Position().Height), op.Position()))
	op.Left.Accept(v)
	op.Right.Accept(v)
	return nil
}

func (v *PrintVisitor) VisitPartialInnerHashOp(op PartialInnerHashOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sPartialInnerHashOp(%v)", v.indent(op.Position().Height), op.Position()))
	op.Left.Accept(v)
	return nil
}

func (v *PrintVisitor) VisitGetCacheOp(op GetCacheOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sGetCacheOp(%v)", v.indent(op.Position().Height), op.Position()))
	return nil
}

func (v *PrintVisitor) VisitPutCacheOp(op PutCacheOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sPutCacheOp(%v)", v.indent(op.Position().Height), op.Position()))
	return op.Operation.Accept(v)
}

func (v *PrintVisitor) VisitMutateOp(op MutateOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sMutateOp(%v)", v.indent(op.Position().Height), op.Position()))
	return op.Operation.Accept(v)
}

func (v *PrintVisitor) VisitCollectOp(op CollectOp) hashing.Digest {
	v.tokens = append(v.tokens, fmt.Sprintf("%sCollectOp(%v)", v.indent(op.Position().Height), op.Position()))
	return op.Operation.Accept(v)
}

func (v PrintVisitor) indent(height uint16) string {
	indents := make([]string, 0)
	for i := height; i < v.height; i++ {
		indents = append(indents, "\t")
	}
	return strings.Join(indents, "")
}
