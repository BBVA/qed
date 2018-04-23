// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package balloon

import (
// "verifiabledata/balloon/history"
)

type Verificator struct{}

func NewVerificator() *Verificator {
	return &Verificator{}
}

func (v *Verificator) Verify(proof *MembershipProof) (bool, error) {
	return false, nil
}
