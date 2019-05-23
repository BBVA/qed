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

package crypto

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ed25519"
)

func NewEd25519KeyPairFile(path string) {
	// Generate a new private/public keypair for QED server
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
	err := ioutil.WriteFile(fmt.Sprintf("%s/id_ed25519", path), privKey, 0644)
	if err != nil {
		panic(err)
	}
	_ = ioutil.WriteFile(fmt.Sprintf("%s/id_ed25519.pub", path), pubKey, 0644)

	fmt.Printf("New keypair generated => %v/id_ed25519|.pub\n", path)
}
