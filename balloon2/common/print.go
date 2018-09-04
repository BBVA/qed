package common

import (
	"fmt"
	"strings"
)

type PrintVisitor struct {
	tokens []string
	height uint16
}

func NewPrintVisitor(height uint16) *PrintVisitor {
	return &PrintVisitor{tokens: make([]string, 1), height: height}
}

func (v PrintVisitor) Result() string {
	return fmt.Sprintf("\n%s", strings.Join(v.tokens[:], "\n"))
}

func (v *PrintVisitor) VisitRoot(pos Position) {
	v.tokens = append(v.tokens, fmt.Sprintf("Root(%v)", pos))
}

func (v *PrintVisitor) VisitNode(pos Position) {
	v.tokens = append(v.tokens, fmt.Sprintf("%sNode(%v)", v.indent(pos.Height()), pos))
}

func (v *PrintVisitor) VisitPartialNode(pos Position) {
	v.tokens = append(v.tokens, fmt.Sprintf("%sPartialNode(%v)", v.indent(pos.Height()), pos))
}
func (v *PrintVisitor) VisitLeaf(pos Position, value []byte) {
	v.tokens = append(v.tokens, fmt.Sprintf("%sLeaf(%v)[%x]", v.indent(pos.Height()), pos, value))
}
func (v *PrintVisitor) VisitCached(pos Position, cachedDigest Digest) {
	v.tokens = append(v.tokens, fmt.Sprintf("%sCached(%v)[%x]", v.indent(pos.Height()), pos, cachedDigest))
}
func (v *PrintVisitor) VisitCollectable(pos Position) {
	v.tokens = append(v.tokens, fmt.Sprintf("%Collectable(%v)", v.indent(pos.Height()), pos))
}

func (v PrintVisitor) indent(height uint16) string {
	indents := make([]string, 0)
	for i := height; i < v.height; i++ {
		indents = append(indents, "\t")
	}
	return strings.Join(indents, "")
}
