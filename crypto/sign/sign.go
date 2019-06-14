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

// Package sign implements funcionality to create signers, which are
// able to sign messages and verify signed messages.
package sign

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ed25519"
)

// Signer is the interface implemented by any value that has Sign and Verify methods.
// Signers are able to sign messages and verify them using a signature.
type Signer interface {
	Sign(message []byte) ([]byte, error)
	Verify(message, sig []byte) (bool, error)
}

type Ed25519Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewEd25519Signer creates an ed25519 signer from scratch.
func NewEd25519Signer() Signer {

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	return &Ed25519Signer{
		privateKey,
		publicKey,
	}

}

// NewEd25519SignerFromFile creates an ed25519 signer using existing private and
// public keys. It also checks that keys are usable.
func NewEd25519SignerFromFile(privateKeyPath string) (Signer, error) {

	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	publicKeyBytes, err := ioutil.ReadFile(fmt.Sprintf("%v.pub", privateKeyPath))
	if err != nil {
		return nil, err
	}

	signer := &Ed25519Signer{
		privateKeyBytes,
		publicKeyBytes,
	}

	message := []byte("test message")
	sig, _ := signer.Sign(message)
	result, _ := signer.Verify(message, sig)
	if result != true {
		return nil, errors.New("key is unusable")
	}

	return signer, nil

}

func (s *Ed25519Signer) Sign(message []byte) ([]byte, error) {
	return ed25519.Sign(s.privateKey, message), nil
}

func (s *Ed25519Signer) Verify(message, sig []byte) (bool, error) {
	return ed25519.Verify(s.publicKey, message, sig), nil
}
