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

package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
)

func CreateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	return key, &key.PublicKey, nil
}

func CreateKeyPairPEM() ([]byte, []byte, error) {
	priv, pub, err := CreateKeyPair()
	if err != nil {
		return nil, nil, err
	}
	privPEM, err := privateKeyAsPEM(priv)
	if err != nil {
		return nil, nil, err
	}
	pubPEM, err := publicKeyAsPEM(pub)
	if err != nil {
		return nil, nil, err
	}
	return privPEM, pubPEM, nil
}

func CreateKeyPairFiles(directory, name string) (string, string, error) {
	priv, pub, err := CreateKeyPair()
	if err != nil {
		return "", "", err
	}
	privPEM, err := privateKeyAsPEM(priv)
	if err != nil {
		return "", "", err
	}
	pubPEM, err := publicKeyAsPEM(pub)
	if err != nil {
		return "", "", err
	}
	privFile, err := SavePEMToFile(directory+"/name", privPEM)
	if err != nil {
		return "", "", err
	}
	pubFile, err := SavePEMToFile(directory+"/name.pub", pubPEM)
	if err != nil {
		return "", "", err
	}
	return privFile, pubFile, nil
}

func privateKeyAsPEM(key *rsa.PrivateKey) ([]byte, error) {
	keyPEM := new(bytes.Buffer)
	if err := pem.Encode(keyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return nil, err
	}
	return keyPEM.Bytes(), nil
}

func publicKeyAsPEM(key *rsa.PublicKey) ([]byte, error) {
	keyPEM := new(bytes.Buffer)
	if err := pem.Encode(keyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key),
	}); err != nil {
		return nil, err
	}
	return keyPEM.Bytes(), nil
}

func certAsPEM(cert []byte) ([]byte, error) {
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}); err != nil {
		return nil, err
	}
	return certPEM.Bytes(), nil
}

func SavePEMToFile(path string, pem []byte) (string, error) {

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	_, err = f.Write(pem)
	if err != nil {
		f.Close()
		return "", err
	}

	if err = f.Close(); err != nil {
		return "", nil
	}

	return f.Name(), nil
}

func CreateCACertificate(subject pkix.Name, expire time.Time, pubKey, privKey []byte) ([]byte, error) {

	var err error
	var pub *rsa.PublicKey
	var priv *rsa.PrivateKey
	pubBlock, _ := pem.Decode(pubKey)
	if pubBlock == nil {
		return nil, errors.New("private key pem decoding failed")
	}
	pub, err = x509.ParsePKCS1PublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse public key")
	}
	privBlock, _ := pem.Decode(privKey)
	if privBlock == nil {
		return nil, errors.New("public key pem decoding failed")
	}
	priv, err = x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse private key")
	}

	certTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(2019),
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              expire,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.SHA512WithRSA,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, pub, priv)
	if err != nil {
		return nil, err
	}

	certPEM, err := certAsPEM(certBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to encode certificate as PEM")
	}

	return certPEM, nil
}

func CreateCertificate(subject pkix.Name, expire time.Time, host string, parentCert, pubKey, privKey []byte) ([]byte, error) {

	var parentTemplate *x509.Certificate
	var err error
	if parentCert != nil {
		parentBlock, _ := pem.Decode(parentCert)
		if parentBlock == nil {
			return nil, errors.New("parent certificate pem decoding failed")
		}
		parentTemplate, err = x509.ParseCertificate(parentBlock.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse parent certificate")
		}
	}

	var pub *rsa.PublicKey
	var priv *rsa.PrivateKey
	pubBlock, _ := pem.Decode(pubKey)
	if pubBlock == nil {
		return nil, errors.New("private key pem decoding failed")
	}
	pub, err = x509.ParsePKCS1PublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse public key")
	}
	privBlock, _ := pem.Decode(privKey)
	if privBlock == nil {
		return nil, errors.New("public key pem decoding failed")
	}
	priv, err = x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse private key")
	}

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     expire,
		IsCA:         false,
		SubjectKeyId: []byte{1, 2, 3, 4, 5, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	ip := net.ParseIP(host)
	// Check if host parameter is DNSName or IPAddr
	if ip == nil {
		certTemplate.DNSNames = append(certTemplate.DNSNames, host)
	} else {
		certTemplate.IPAddresses = append(certTemplate.IPAddresses, ip)
	}

	var certBytes []byte
	if parentTemplate == nil {
		certBytes, err = x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, pub, priv)
	} else {
		certBytes, err = x509.CreateCertificate(rand.Reader, certTemplate, parentTemplate, pub, priv)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create certificate")
	}

	certPEM, err := certAsPEM(certBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to encode certificate as PEM")
	}

	return certPEM, nil
}

func CreateCsr(subject pkix.Name, expire time.Time, host string, priv *rsa.PrivateKey) (*x509.CertificateRequest, []byte, error) {

	cert := &x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA512WithRSA,
		Subject:            subject,
		ExtraExtensions: []pkix.Extension{
			{
				Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
				Critical: true,
			},
		},
	}

	ip := net.ParseIP(host)
	// Check if host parameter is DNSName or IPAddr
	if ip == nil {
		cert.DNSNames = append(cert.DNSNames, host)
	} else {
		cert.IPAddresses = append(cert.IPAddresses, ip)
	}

	certBytes, err := x509.CreateCertificateRequest(rand.Reader, cert, priv)
	if err != nil {
		return nil, nil, err
	}

	return cert, certBytes, err
}

func SignCsr(csrPem, caPublicKeyPem, caPrivateKeyPem []byte, expire time.Time) ([]byte, error) {

	// get ca public key
	caPublicKeyBlock, _ := pem.Decode(caPublicKeyPem)
	if caPublicKeyBlock == nil {
		return nil, fmt.Errorf("pem decode failed")
	}
	caPublicKey, err := x509.ParseCertificate(caPublicKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	// get ca private key
	caPrivateKeyBlock, _ := pem.Decode(caPrivateKeyPem)
	if caPrivateKeyBlock == nil {
		return nil, fmt.Errorf("pem decode failed")
	}
	caPrivateKey, err := x509.ParsePKCS1PrivateKey(caPrivateKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	// get client certificate request
	csrBlock, _ := pem.Decode(csrPem)
	if csrBlock == nil {
		return nil, fmt.Errorf("pem decode failed")
	}
	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return nil, err
	}

	if err := csr.CheckSignature(); err != nil {
		return nil, errors.Wrap(err, "unable to check signature")
	}

	// create client certificate template
	crtTemplate := x509.Certificate{
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,
		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,
		SerialNumber:       big.NewInt(1658),
		Issuer:             caPublicKey.Subject,
		IsCA:               false,
		NotBefore:          time.Now(),
		NotAfter:           time.Now().Add(expire.Sub(time.Now())),
		KeyUsage:           x509.KeyUsageDigitalSignature,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}

	crt, err := x509.CreateCertificate(rand.Reader, &crtTemplate, caPublicKey, csr.PublicKey, caPrivateKey)
	if err != nil {
		return nil, err
	}

	return crt, nil
}
