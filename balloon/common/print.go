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

package common

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

func (v *PrintVisitor) VisitCached(pos Position, cachedDigest hashing.Digest) {
	v.tokens = append(v.tokens, fmt.Sprintf("%sCached(%v)[%x]", v.indent(pos.Height()), pos, cachedDigest))
}

func (v *PrintVisitor) VisitCollectable(pos Position) {
	v.tokens = append(v.tokens, fmt.Sprintf("%sCollectable(%v)", v.indent(pos.Height()), pos))
}

func (v *PrintVisitor) VisitCacheable(pos Position) {
	v.tokens = append(v.tokens, fmt.Sprintf("%sCacheable(%v)", v.indent(pos.Height()), pos))
}

func (v PrintVisitor) indent(height uint16) string {
	indents := make([]string, 0)
	for i := height; i < v.height; i++ {
		indents = append(indents, "\t")
	}
	return strings.Join(indents, "")
}
