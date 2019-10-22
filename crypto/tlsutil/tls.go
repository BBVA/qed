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
package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sync"
)

// Config is the config used to create a tls.Config
type Config struct {
	// UseTLS is used to enable TLS for outgoing connections to other TLS-capable QED
	// servers. This doesn't imply any verification, it only enables TLS if possible.
	UseTLS bool

	// VerifyIncoming is used to verify the authenticity of incoming connections.
	// This means that TCP requests are forbidden, only allowing for TLS. TLS connecitons
	// must match a provided certificate authority. This is used to verify authenticity
	// of other QED nodes trying to connect to us.
	VerifyIncoming bool

	// VerifyOutgoing is used to force verification of the authenticity of outgoing connections.
	// This means that TLS requests are used, and TCP requests are forbidden. TLS connections
	// must match a provided certificate authority. This is used to verify authenticity of
	// other QED nodes we are connecting to.
	VerifyOutgoing bool

	// VerifyServerHostname is used to enable hostname verification of servers.
	VerifyServerHostname bool

	// CAFilePath is a path to a certificate authority file. This is used with VerifyIncoming
	// or VerifyOutgoing to verify the TLS connection.
	CAFilePath string

	// CertFilePath is a path to a TLS certificate that must be provided to serve TLS connections.
	CertFilePath string

	// KeyFilePath is a path to a TLS key that must be provided to serve TLS connections.
	KeyFilePath string
}

// KeyPair is used to open and parse a certificate and key pair.
func (c *Config) KeyPair() (*tls.Certificate, error) {
	if c.CertFilePath == "" || c.KeyFilePath == "" {
		return nil, nil
	}
	cert, err := tls.LoadX509KeyPair(c.CertFilePath, c.KeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load cert/key pair: %v", err)
	}
	return &cert, err
}

// TLSConfigurator holds a config and is responsible for generating
// all *tls.Config.
type TLSConfigurator struct {
	sync.RWMutex
	conf *Config
}

// NewTLSConfigurator creates a new TLSConfigurator and sets the provided
// configuration.
func NewTLSConfigurator(config *Config) *TLSConfigurator {
	if config == nil {
		config = &Config{}
	}
	return &TLSConfigurator{conf: config}
}

func (c *TLSConfigurator) AppendCAToPool(pool *x509.CertPool) error {
	if c.conf.CAFilePath == "" {
		return nil
	}
	asn1Data, err := ioutil.ReadFile(c.conf.CAFilePath)
	if err != nil {
		return fmt.Errorf("Failed to read CA file: %v", err)
	}
	if !pool.AppendCertsFromPEM([]byte(asn1Data)) {
		return fmt.Errorf("failed to parse CA certificate in %q", c.conf.CAFilePath)
	}
	return nil
}

// OutgoingTLSConfig generates a TLS configuration for outgoing requests.
// It will return a nil config if this configuration should
// not use TLS for outgoing connections.
func (c *TLSConfigurator) OutgoingTLSConfig() (*tls.Config, error) {

	c.RLock()
	defer c.RUnlock()

	// if VerifyServerHostname is true, that implies VerifyOutgoing
	if c.conf.VerifyServerHostname {
		c.conf.VerifyOutgoing = true
	}

	if !c.conf.UseTLS && !c.conf.VerifyOutgoing {
		return nil, nil
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	if c.conf.VerifyServerHostname {
		conf.InsecureSkipVerify = false
	}

	if c.conf.VerifyOutgoing && c.conf.CAFilePath == "" {
		return nil, fmt.Errorf("VerifyOutgoing set, and no CA certificate provided!")
	}

	// Parse CA certs if any
	conf.RootCAs = x509.NewCertPool()
	if err := c.AppendCAToPool(conf.RootCAs); err != nil {
		return nil, err
	}

	// Add cert/key
	cert, err := c.conf.KeyPair()
	if err != nil {
		return nil, err
	} else if cert != nil {
		conf.Certificates = []tls.Certificate{*cert}
	}

	conf.BuildNameToCertificate()

	return conf, nil
}

// IncomingTLSConfig generates a TLS configuration for incoming requests.
// It will return a nil config if this configuration should
// not use TLS for incoming connections.
func (c *TLSConfigurator) IncomingTLSConfig() (*tls.Config, error) {

	c.RLock()
	defer c.RUnlock()

	if !c.conf.UseTLS {
		return nil, nil
	}

	conf := &tls.Config{
		ClientCAs:  x509.NewCertPool(),
		ClientAuth: tls.NoClientCert,
	}

	// Parse CA certs if any
	if c.conf.CAFilePath != "" {
		if err := c.AppendCAToPool(conf.ClientCAs); err != nil {
			return nil, err
		}
	}

	// Add cert/key
	cert, err := c.conf.KeyPair()
	if err != nil {
		return nil, err
	} else if cert != nil {
		conf.Certificates = []tls.Certificate{*cert}
	}

	if c.conf.VerifyIncoming {
		conf.ClientAuth = tls.RequireAndVerifyClientCert
		conf.PreferServerCipherSuites = true
		if c.conf.CAFilePath == "" {
			return nil, fmt.Errorf("VerifyIncoming set, and no CA certificate provided!")
		}
		if cert == nil {
			return nil, fmt.Errorf("VerifyIncoming set, and no Cert/Key pair provided!")
		}
	}

	conf.BuildNameToCertificate()

	return conf, nil
}
