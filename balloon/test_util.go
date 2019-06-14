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

package balloon

import (
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/crypto/hashing"
)

func NewFakeQueryProof(shouldVerify bool, value []byte, hasher hashing.Hasher) *hyper.QueryProof {
	if shouldVerify {
		return hyper.NewQueryProof([]byte{0}, value, hyper.AuditPath{"128|7": value}, hasher)
	}
	return hyper.NewQueryProof([]byte{0}, []byte{0}, hyper.AuditPath{}, hasher)
}

func NewFakeMembershipProof(shouldVerify bool, hasher hashing.Hasher) *history.MembershipProof {
	if shouldVerify {
		return history.NewMembershipProof(0, 0, history.AuditPath{}, hasher)
	}
	return history.NewMembershipProof(1, 1, history.AuditPath{}, hasher)
}
