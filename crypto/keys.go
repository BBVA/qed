/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, n.A.
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

// Package crypto implements key generators.
package crypto

import (
	"crypto/rand"
	"io/ioutil"

	"golang.org/x/crypto/ed25519"
)

// NewEd25519SignerKeysFile generates a new private/public signer key.
// Input parameter is the full path to the output directory where the keys
// will be stored. The function output is the full path to our new signer keys
// and an error. Eg: (/var/tmp/qed_ed25519, /var/tmp/qed_ed25519.pub, nil)
func NewEd25519SignerKeysFile(path string) (string, string, error) {
	outPriv := path + "/qed_ed25519"
	outPub := outPriv + ".pub"

	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)

	err := ioutil.WriteFile(outPriv, privKey, 0644)
	if err != nil {
		return outPub, outPriv, err
	}
	_ = ioutil.WriteFile(outPub, pubKey, 0644)

	return outPub, outPriv, nil
}
