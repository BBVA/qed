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

package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

// NewSelfSignedCert generates a new Tls certificate and private key.
// First parameter is the full path to the output directory where the
// certificate and the key will be stored. The second one is the host
// (DNSName or IPAddr) for which we are signing the certificate.
// The function output is the full path to our new certificate and
// private key. Eg: (/var/tmp/qed_key.pem, /var/tmp/qed_cert.pem, nil)
func NewSelfSignedCert(path, host string) (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(1 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", err
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

	ip := net.ParseIP(host)
	// Check if host parameter is DNSName or IPAddr
	if ip == nil {
		template.DNSNames = append(template.DNSNames, host)
	} else {
		template.IPAddresses = append(template.IPAddresses, ip)
	}
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	certOut, err := os.Create(path + "/qed_cert.pem")
	if err != nil {
		return "", "", err
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", "", err
	}

	if err := certOut.Close(); err != nil {
		return "", "", err
	}

	keyOut, err := os.OpenFile(path+"/qed_key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", err
	}
	keyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}
	if err := pem.Encode(keyOut, keyBlock); err != nil {
		return "", "", err
	}

	if err := keyOut.Close(); err != nil {
		return "", "", err
	}

	return certOut.Name(), keyOut.Name(), nil
}

// NewCsrRequest generates a private key and Certificate Signing Request(CSR).
// This enables us to use custom CA to sign the server certificates.
// First parameter is the full path to the output directory where the
// CSR and the key will be stored. The second one is the host
// (DNSName or IPAddr) for which we are creating the request.
// The function output is the full path to our new CSR and private key.
// Eg: (/var/tmp/qed.csr, /var/tmp/qed_key.pem, nil)
func NewCsrRequest(path, host string) (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(1 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", err
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

	ip := net.ParseIP(host)
	// Check if host parameter is DNSName or IPAddr
	if ip == nil {
		template.DNSNames = append(template.DNSNames, host)
	} else {
		template.IPAddresses = append(template.IPAddresses, ip)
	}
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	var csrTemplate = x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		SignatureAlgorithm: x509.SHA512WithRSA,
		ExtraExtensions: []pkix.Extension{
			{
				Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
				Critical: true,
			},
		},
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, priv)
	if err != nil {
		return "", "", err
	}

	csrBlock := &pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	}

	csrOut, err := os.OpenFile(path+"/qed.csr", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", err
	}

	if err := pem.Encode(csrOut, csrBlock); err != nil {
		return "", "", err
	}

	keyOut, err := os.OpenFile(path+"/qed_key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", err
	}

	keyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}
	if err := pem.Encode(keyOut, keyBlock); err != nil {
		return "", "", err
	}

	if err := keyOut.Close(); err != nil {
		return "", "", err
	}

	return csrOut.Name(), keyOut.Name(), nil
}
