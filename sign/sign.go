package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/asn1"
	"fmt"
	"io"
	"math/big"
	"os"

	"github.com/bbva/qed/log"
	"golang.org/x/crypto/ed25519"
)

type Signable interface {
	Sign(message []byte) ([]byte, error)
	Verify(message, sig []byte) (bool, error)
}

var std Signable = NewEdSigner()

func Sign(message []byte) ([]byte, error)      { return std.Sign(message) }
func Verify(message, sig []byte) (bool, error) { return std.Verify(message, sig) }

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

type EdSigner struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewEdSigner() Signable {

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	return &EdSigner{
		privateKey,
		publicKey,
	}

}

func (s *EdSigner) Sign(message []byte) ([]byte, error) {
	return ed25519.Sign(s.privateKey, message), nil
}

func (s *EdSigner) Verify(message, sig []byte) (bool, error) {
	return ed25519.Verify(s.publicKey, message, sig), nil
}

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
