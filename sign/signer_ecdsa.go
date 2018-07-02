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
package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"io"
	"math/big"
)

type EcdsaSigner struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	rng        io.Reader
	curve      elliptic.Curve
}

type ecdsaSignature struct {
	R, S *big.Int
}

func NewEcdsaSigner() Signable {

	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}

	return &EcdsaSigner{
		privateKey,
		&privateKey.PublicKey,
		rand.Reader,
		curve,
	}

}

func (e *EcdsaSigner) Sign(message []byte) ([]byte, error) {

	r, s, err := ecdsa.Sign(e.rng, e.privateKey, message)
	if err != nil {
		return nil, err
	}

	return asn1.Marshal(ecdsaSignature{r, s})
}

func (s *EcdsaSigner) Verify(message, sig []byte) (bool, error) {

	ecdsaSig := &ecdsaSignature{}
	_, err := asn1.Unmarshal(sig, ecdsaSig)
	if err != nil {
		return false, err
	}

	return ecdsa.Verify(s.publicKey, message, ecdsaSig.R, ecdsaSig.S), nil

}
