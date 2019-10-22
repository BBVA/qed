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

package consensus

import (
	"crypto/x509/pkix"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/bbva/qed/crypto/tlsutil"
	"github.com/bbva/qed/log"
	test_tls "github.com/bbva/qed/testutils/tls"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestCreateTransportAndClose(t *testing.T) {
	tn, err := NewCMuxTCPTransport("127.0.0.1:0", 3, 10*time.Second, nil, fakeRegister)
	require.NoError(t, err, "failed to create transport")
	require.Equal(t, "127.0.0.1:0", string(tn.LocalAddr()), "transport address set incorrectly")
	require.NoError(t, tn.Close(), "failed to close transport")
}

func TestTransportBadAddress(t *testing.T) {
	_, err := NewCMuxTCPTransport("0.0.0.0:0", 3, 10*time.Second, nil, fakeRegister)
	require.EqualError(t, err, errNotAdvertisable.Error())
}

func TestLayerDial(t *testing.T) {

	done := make(chan bool)
	var layer1, layer2 *CMuxTCPStreamLayer
	var err error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		layer1, err = NewCMuxTCPStreamLayer("127.0.0.1:9000", nil, fakeRegister)
		assert.NoError(t, err)
		wg.Done()
		if layer1 != nil {
			defer layer1.Close()
		}
		<-done
	}()

	go func() {
		layer2, err = NewCMuxTCPStreamLayer("127.0.0.1:9001", nil, fakeRegister)
		assert.NoError(t, err)
		wg.Done()
		if layer2 != nil {
			defer layer2.Close()
		}
		<-done
	}()

	wg.Wait()
	require.NotNil(t, layer1)
	conn, err := layer1.Dial(raft.ServerAddress(layer2.Addr().String()), 10*time.Second)
	require.NoError(t, err)

	require.Equal(t, layer2.Addr().String(), conn.RemoteAddr().String())
	conn.Close()
	close(done)

}

func TestMutualTLSWithoutVerifying(t *testing.T) {

	certsDirPath, cleanF, err := createCertsDir(t.Name())
	require.NoError(t, err)
	defer cleanF()

	// create key pair
	priv, pub, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)
	keyPath, err := test_tls.SavePEMToFile(certsDirPath+"/test", priv)
	require.NoError(t, err)

	// create self signed cert
	subject := pkix.Name{
		Organization: []string{""},
	}
	cert, err := test_tls.CreateCertificate(subject, time.Now().Add(24*time.Hour), "127.0.0.1", nil, pub, priv)
	require.NoError(t, err)
	certPath, err := test_tls.SavePEMToFile(certsDirPath+"/test.crt", cert)
	require.NoError(t, err)

	// configure TLS
	tlsConf := &tlsutil.Config{
		UseTLS:               true,
		VerifyIncoming:       false,
		VerifyOutgoing:       false,
		VerifyServerHostname: false,
		CertFilePath:         certPath,
		KeyFilePath:          keyPath,
	}

	// start two layers
	done := make(chan bool)
	defer close(done)
	var layer1, layer2 *CMuxTCPStreamLayer

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		layer1, err = NewCMuxTCPStreamLayer("127.0.0.1:9000", tlsutil.NewTLSConfigurator(tlsConf), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		<-done
	}()
	go func() {
		layer2, err = NewCMuxTCPStreamLayer("127.0.0.1:9001", tlsutil.NewTLSConfigurator(tlsConf), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		for {
			conn, err := layer2.Accept()
			if err != nil {
				log.L().Errorf("server: accept: %s", err)
				break
			}
			defer conn.Close()
			log.L().Debugf("server: accepted from %s\n", conn.RemoteAddr())
			handleClientConnection(conn)
			log.L().Debug("server: conn: closed")
		}
		<-done
	}()
	wg.Wait()
	require.NotNil(t, layer1)
	defer layer1.Close()
	require.NotNil(t, layer2)
	defer layer2.Close()

	// try to dial from layer 1 to layer 2
	conn, err := layer1.Dial(raft.ServerAddress(layer2.Addr().String()), 10*time.Second)
	require.NoError(t, err)
	defer conn.Close()

	// write and read
	require.NoError(t, writeAndRead(conn))

}

func TestMutualTLSVerifyingWithWrongBothCA(t *testing.T) {

	certsDirPath, cleanF, err := createCertsDir(t.Name())
	require.NoError(t, err)
	defer cleanF()

	// create key pairs
	priv1, pub1, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)
	keyPath1, err := test_tls.SavePEMToFile(certsDirPath+"/test1", priv1)
	require.NoError(t, err)
	priv2, pub2, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)
	keyPath2, err := test_tls.SavePEMToFile(certsDirPath+"/test2", priv2)
	require.NoError(t, err)

	// create two self signed cert
	subject := pkix.Name{
		Organization: []string{""},
	}
	cert1, err := test_tls.CreateCertificate(subject, time.Now().Add(24*time.Hour), "127.0.0.1", nil, pub1, priv1)
	require.NoError(t, err)
	cert1Path, err := test_tls.SavePEMToFile(certsDirPath+"/test1.crt", cert1)
	require.NoError(t, err)
	cert2, err := test_tls.CreateCertificate(subject, time.Now().Add(24*time.Hour), "127.0.0.1", nil, pub2, priv2)
	require.NoError(t, err)
	cert2Path, err := test_tls.SavePEMToFile(certsDirPath+"/test2.crt", cert2)
	require.NoError(t, err)

	// create TLS
	tlsConf1 := &tlsutil.Config{
		UseTLS:               true,
		VerifyIncoming:       true,
		VerifyOutgoing:       true,
		VerifyServerHostname: false,
		CertFilePath:         cert1Path,
		KeyFilePath:          keyPath1,
		CAFilePath:           cert1Path, // self-signed
	}
	tlsConf2 := &tlsutil.Config{
		UseTLS:               true,
		VerifyIncoming:       true,
		VerifyOutgoing:       true,
		VerifyServerHostname: false,
		CertFilePath:         cert2Path,
		KeyFilePath:          keyPath2,
		CAFilePath:           cert2Path, // self-signed
	}

	done := make(chan bool)
	defer close(done)
	var layer1, layer2 *CMuxTCPStreamLayer

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		layer1, err = NewCMuxTCPStreamLayer("127.0.0.1:9000", tlsutil.NewTLSConfigurator(tlsConf1), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		<-done
	}()

	go func() {
		layer2, err = NewCMuxTCPStreamLayer("127.0.0.1:9001", tlsutil.NewTLSConfigurator(tlsConf2), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		for {
			conn, err := layer2.Accept()
			if err != nil {
				log.L().Errorf("server: accept: %s", err)
				break
			}
			defer conn.Close()
			log.L().Debugf("server: accepted from %s\n", conn.RemoteAddr())
			require.Error(t, handleClientConnection(conn))
			log.L().Debug("server: conn: closed")
		}
		<-done
	}()

	wg.Wait()
	require.NotNil(t, layer1)
	defer layer1.Close()
	require.NotNil(t, layer2)
	defer layer2.Close()

	// try to dial from layer 1 to layer 2
	conn, err := layer1.Dial(raft.ServerAddress(layer2.Addr().String()), 10*time.Second)
	require.NoError(t, err)
	defer conn.Close()

	// write and read
	require.Error(t, writeAndRead(conn))
}

func TestMutualTLSVerifyingWithWrongServerCA(t *testing.T) {

	certsDirPath, cleanF, err := createCertsDir(t.Name())
	require.NoError(t, err)
	defer cleanF()

	// create CA key pair
	caPriv, caPub, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)

	// create key pairs
	priv1, pub1, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)
	keyPath1, err := test_tls.SavePEMToFile(certsDirPath+"/test1", priv1)
	require.NoError(t, err)
	priv2, pub2, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)
	keyPath2, err := test_tls.SavePEMToFile(certsDirPath+"/test2", priv2)
	require.NoError(t, err)

	subject := pkix.Name{
		Organization: []string{""},
	}

	// create CA cert
	caCert, err := test_tls.CreateCACertificate(subject, time.Now().Add(24*time.Hour), caPub, caPriv)
	require.NoError(t, err)
	caCertPath, err := test_tls.SavePEMToFile(certsDirPath+"/ca.crt", caCert)
	require.NoError(t, err)

	// create other
	cert1, err := test_tls.CreateCertificate(subject, time.Now().Add(24*time.Hour), "127.0.0.1", caCert, pub1, caPriv)
	require.NoError(t, err)
	cert1Path, err := test_tls.SavePEMToFile(certsDirPath+"/test1.crt", cert1)
	require.NoError(t, err)
	cert2, err := test_tls.CreateCertificate(subject, time.Now().Add(24*time.Hour), "127.0.0.1", nil, pub2, priv2)
	require.NoError(t, err)
	cert2Path, err := test_tls.SavePEMToFile(certsDirPath+"/test2.crt", cert2)
	require.NoError(t, err)

	// create TLS
	tlsConf1 := &tlsutil.Config{
		UseTLS:               true,
		VerifyIncoming:       true,
		VerifyOutgoing:       true,
		VerifyServerHostname: true,
		CertFilePath:         cert1Path,
		KeyFilePath:          keyPath1,
		CAFilePath:           caCertPath,
	}
	tlsConf2 := &tlsutil.Config{
		UseTLS:               true,
		VerifyIncoming:       true,
		VerifyOutgoing:       true,
		VerifyServerHostname: true,
		CertFilePath:         cert2Path,
		KeyFilePath:          keyPath2,
		CAFilePath:           caCertPath,
	}

	done := make(chan bool)
	defer close(done)
	var layer1, layer2 *CMuxTCPStreamLayer

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		layer1, err = NewCMuxTCPStreamLayer("127.0.0.1:9000", tlsutil.NewTLSConfigurator(tlsConf1), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		<-done
	}()

	go func() {
		layer2, err = NewCMuxTCPStreamLayer("127.0.0.1:9001", tlsutil.NewTLSConfigurator(tlsConf2), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		for {
			conn, err := layer2.Accept()
			if err != nil {
				log.L().Errorf("server: accept: %s", err)
				break
			}
			defer conn.Close()
			log.L().Debugf("server: accepted from %s\n", conn.RemoteAddr())
			require.Error(t, handleClientConnection(conn))
			log.L().Debug("server: conn: closed")
		}
		<-done
	}()

	wg.Wait()
	require.NotNil(t, layer1)
	defer layer1.Close()
	require.NotNil(t, layer2)
	defer layer2.Close()

	// try to dial from layer 1 to layer 2
	_, err = layer1.Dial(raft.ServerAddress(layer2.Addr().String()), 10*time.Second)
	require.Error(t, err)

}

func TestMutualTLSVerifyingWithWrongClientCA(t *testing.T) {

	certsDirPath, cleanF, err := createCertsDir(t.Name())
	require.NoError(t, err)
	defer cleanF()

	// create CA key pair
	caPriv, caPub, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)

	// create key pairs
	priv1, pub1, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)
	keyPath1, err := test_tls.SavePEMToFile(certsDirPath+"/test1", priv1)
	require.NoError(t, err)
	priv2, pub2, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)
	keyPath2, err := test_tls.SavePEMToFile(certsDirPath+"/test2", priv2)
	require.NoError(t, err)

	subject := pkix.Name{
		Organization: []string{""},
	}

	// create CA cert
	caCert, err := test_tls.CreateCACertificate(subject, time.Now().Add(24*time.Hour), caPub, caPriv)
	require.NoError(t, err)
	caCertPath, err := test_tls.SavePEMToFile(certsDirPath+"/ca.crt", caCert)
	require.NoError(t, err)

	// create other
	cert1, err := test_tls.CreateCertificate(subject, time.Now().Add(24*time.Hour), "127.0.0.1", nil, pub1, priv1)
	require.NoError(t, err)
	cert1Path, err := test_tls.SavePEMToFile(certsDirPath+"/test1.crt", cert1)
	require.NoError(t, err)
	cert2, err := test_tls.CreateCertificate(subject, time.Now().Add(24*time.Hour), "127.0.0.1", caCert, pub2, caPriv)
	require.NoError(t, err)
	cert2Path, err := test_tls.SavePEMToFile(certsDirPath+"/test2.crt", cert2)
	require.NoError(t, err)

	// create TLS
	tlsConf1 := &tlsutil.Config{
		UseTLS:               true,
		VerifyIncoming:       true,
		VerifyOutgoing:       true,
		VerifyServerHostname: true,
		CertFilePath:         cert1Path,
		KeyFilePath:          keyPath1,
		CAFilePath:           caCertPath,
	}
	tlsConf2 := &tlsutil.Config{
		UseTLS:               true,
		VerifyIncoming:       true,
		VerifyOutgoing:       true,
		VerifyServerHostname: true,
		CertFilePath:         cert2Path,
		KeyFilePath:          keyPath2,
		CAFilePath:           caCertPath,
	}

	done := make(chan bool)
	defer close(done)
	var layer1, layer2 *CMuxTCPStreamLayer

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		layer1, err = NewCMuxTCPStreamLayer("127.0.0.1:9000", tlsutil.NewTLSConfigurator(tlsConf1), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		<-done
	}()

	go func() {
		layer2, err = NewCMuxTCPStreamLayer("127.0.0.1:9001", tlsutil.NewTLSConfigurator(tlsConf2), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		for {
			conn, err := layer2.Accept()
			if err != nil {
				log.L().Errorf("server: accept: %s", err)
				break
			}
			defer conn.Close()
			log.L().Debugf("server: accepted from %s\n", conn.RemoteAddr())
			require.Error(t, handleClientConnection(conn))
			log.L().Debug("server: conn: closed")
		}
		<-done
	}()

	wg.Wait()
	require.NotNil(t, layer1)
	defer layer1.Close()
	require.NotNil(t, layer2)
	defer layer2.Close()

	// try to dial from layer 1 to layer 2
	conn, err := layer1.Dial(raft.ServerAddress(layer2.Addr().String()), 10*time.Second)
	require.NoError(t, err)
	defer conn.Close()

	// write and read
	require.Error(t, writeAndRead(conn))
}

func TestMutualTLSVerifying(t *testing.T) {

	certsDirPath, cleanF, err := createCertsDir(t.Name())
	require.NoError(t, err)
	defer cleanF()

	// create CA key pair
	caPriv, caPub, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)

	// create other key pairs
	priv1, pub1, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)
	keyPath1, err := test_tls.SavePEMToFile(certsDirPath+"/test1", priv1)
	require.NoError(t, err)

	priv2, pub2, err := test_tls.CreateKeyPairPEM()
	require.NoError(t, err)
	keyPath2, err := test_tls.SavePEMToFile(certsDirPath+"/test2", priv2)
	require.NoError(t, err)

	subject := pkix.Name{
		Organization: []string{""},
	}

	// create CA cert
	caCert, err := test_tls.CreateCACertificate(subject, time.Now().Add(24*time.Hour), caPub, caPriv)
	require.NoError(t, err)
	caCertPath, err := test_tls.SavePEMToFile(certsDirPath+"/ca.crt", caCert)
	require.NoError(t, err)

	// create CA signed certs
	cert1, err := test_tls.CreateCertificate(subject, time.Now().Add(1*time.Hour), "127.0.0.1", caCert, pub1, caPriv)
	require.NoError(t, err)
	certPath1, err := test_tls.SavePEMToFile(certsDirPath+"/test1.crt", cert1)
	require.NoError(t, err)

	cert2, err := test_tls.CreateCertificate(subject, time.Now().Add(1*time.Hour), "127.0.0.1", caCert, pub2, caPriv)
	require.NoError(t, err)
	certPath2, err := test_tls.SavePEMToFile(certsDirPath+"/test2.crt", cert2)
	require.NoError(t, err)

	// create TLS
	tlsConf1 := &tlsutil.Config{
		UseTLS:               true,
		VerifyIncoming:       true,
		VerifyOutgoing:       true,
		VerifyServerHostname: true,
		CertFilePath:         certPath1,
		KeyFilePath:          keyPath1,
		CAFilePath:           caCertPath,
	}

	tlsConf2 := &tlsutil.Config{
		UseTLS:               true,
		VerifyIncoming:       true,
		VerifyOutgoing:       true,
		VerifyServerHostname: true,
		CertFilePath:         certPath2,
		KeyFilePath:          keyPath2,
		CAFilePath:           caCertPath,
	}

	done := make(chan bool)
	defer close(done)
	var layer1, layer2 *CMuxTCPStreamLayer

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		layer1, err = NewCMuxTCPStreamLayer("127.0.0.1:9000", tlsutil.NewTLSConfigurator(tlsConf1), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		<-done
	}()

	go func() {
		layer2, err = NewCMuxTCPStreamLayer("127.0.0.1:9001", tlsutil.NewTLSConfigurator(tlsConf2), fakeRegister)
		require.NoError(t, err)
		wg.Done()
		for {
			conn, err := layer2.Accept()
			if err != nil {
				log.L().Errorf("server: accept: %s", err)
				break
			}
			defer conn.Close()
			log.L().Debugf("server: accepted from %s\n", conn.RemoteAddr())
			require.NoError(t, handleClientConnection(conn))
			log.L().Debug("server: conn: closed")
		}
		<-done
	}()

	wg.Wait()
	require.NotNil(t, layer1)
	defer layer1.Close()
	require.NotNil(t, layer2)
	defer layer2.Close()

	conn, err := layer1.Dial(raft.ServerAddress(layer2.Addr().String()), 10*time.Second)
	require.NoError(t, err)
	defer conn.Close()

	// write and read
	require.NoError(t, writeAndRead(conn))

}

var fakeRegister = func(*grpc.Server) {}

func createCertsDir(name string) (string, func(), error) {
	certsDirPath, err := ioutil.TempDir("", name)
	if err != nil {
		return "", nil, err
	}
	return certsDirPath, func() {
		err := os.RemoveAll(certsDirPath)
		if err != nil {
			log.L().Errorf("Unable to remove certs dir: %s", err)
		}
	}, nil
}

func handleClientConnection(conn net.Conn) error {
	var buffer = make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		if err != nil {
			log.L().Debugf("server: conn: read: %s\n", err)
		}
		return err
	}
	log.L().Debugf("server: conn: echo %q", string(buffer[:n]))
	n, err = conn.Write(buffer[:n])
	log.L().Debugf("server: conn: wrote %d bytes", n)

	if err != nil {
		log.L().Warnf("server: write: %s\n", err)
		return err
	}

	return nil
}

func writeAndRead(conn net.Conn) error {
	// send a message
	message := "Hello"
	n, err := io.WriteString(conn, message)
	if err != nil {
		log.L().Debugf("client: write: %s", err)
	}
	if err != nil {
		return err
	}
	log.L().Debugf("client: wrote %q (%d bytes)", message, n)

	// read reply
	reply := make([]byte, 256)
	n, err = conn.Read(reply)
	if err != nil {
		return err
	}
	log.L().Debugf("client: read %q (%d bytes)", string(reply[:n]), n)
	log.L().Debug("client: exiting")
	return nil
}
