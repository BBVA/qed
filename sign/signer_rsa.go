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
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"os"

	"github.com/bbva/qed/log"
)

type RSASigner struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	rng        io.Reader
}

func NewRSASigner(keySize int) Signable {

	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		panic(err)
	}

	return &RSASigner{
		// TODO: this should be parameters external from the application.
		// for now it's for PoC porpouses.
		privateKey,
		&privateKey.PublicKey,
		rand.Reader,
	}

}

func (s *RSASigner) Sign(message []byte) ([]byte, error) {

	sig, err := rsa.SignPKCS1v15(s.rng, s.privateKey, 0, message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from Signing: %s\n", err)
		return nil, err
	}

	log.Debugf("Sig: %x\n", sig)
	return sig, nil

}

func (s *RSASigner) Verify(message, sig []byte) (bool, error) {

	err := rsa.VerifyPKCS1v15(s.publicKey, 0, message, sig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from Verifying: %s\n", err)
		return false, err
	}

	return true, nil

}
