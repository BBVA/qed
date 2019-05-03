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

package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"time"
)

func GenerateSignKey(path string) (string, error) {
	/* public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", err
	} */

	err := ioutil.WriteFile(path+"/id_ed25519", []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACCXiy9taNh+avmEAyzfnlCXxYHzgS3wGPCi52AY8qQyTwAAAKA6Sq8iOkqv
IgAAAAtzc2gtZWQyNTUxOQAAACCXiy9taNh+avmEAyzfnlCXxYHzgS3wGPCi52AY8qQyTw
AAAEAzpsL9rtrmKhL3cEHFcKPEvkd8y/QJXeFTtyhgaYfUDpeLL21o2H5q+YQDLN+eUJfF
gfOBLfAY8KLnYBjypDJPAAAAHWdkaWF6bG9AbWFjc2xhY2suc2xhY2t3YXJlLmVz
-----END OPENSSH PRIVATE KEY-----
`), 0644)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(path+"/id_ed25519.pub", []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJeLL21o2H5q+YQDLN+eUJfFgfOBLfAY8KLnYBjypDJP gdiazlo@macslack.slackware.es"), 0644)
	if err != nil {
		return "", err
	}

	return path + "/id_ed25519", nil
}

func GenerateTlsCert(path string) (string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(1 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	ip := net.ParseIP("127.0.0.1")
	template.IPAddresses = append(template.IPAddresses, ip)
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", err
	}

	certOut, err := os.Create(path + "/cert.pem")
	if err != nil {
		return "", err
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", err
	}

	if err := certOut.Close(); err != nil {
		return "", err
	}

	keyOut, err := os.OpenFile(path+"/key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}
	if err := pem.Encode(keyOut, block); err != nil {
		return "", err
	}

	if err := keyOut.Close(); err != nil {
		return "", err
	}

	return path, nil
}
