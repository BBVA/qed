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
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bbva/qed/crypto/tlsutil"
	"github.com/bbva/qed/log"
	"github.com/hashicorp/raft"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

var (
	errNotAdvertisable = errors.New("local bind address is not advertisable")
	errNotTCP          = errors.New("local address is not a TCP address")
)

// CMuxTCPStreamLayer implements StreamLayer interface for plain TCP.
type CMuxTCPStreamLayer struct {
	advertise net.Addr

	mux        cmux.CMux
	grpcServer *grpc.Server
	listener   net.Listener
	tlsConfig  *tlsutil.TLSConfigurator

	// Tracks if we are closed
	closed    bool
	closeLock sync.Mutex
}

// GRPCServiceRegister is a registering function for GRPC services.
type GRPCServiceRegister func(*grpc.Server)

// NewCMuxTCPStreamLayer creates a CMuxTCPStreamLayer with the given parameters
func NewCMuxTCPStreamLayer(bindAddr string, tlsC *tlsutil.TLSConfigurator, grpcServiceRegister GRPCServiceRegister) (*CMuxTCPStreamLayer, error) {

	advertiseAddr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	// Get TLS config
	if tlsC == nil {
		tlsC = tlsutil.NewTLSConfigurator(&tlsutil.Config{})
	}
	tlsConf, err := tlsC.IncomingTLSConfig()
	if err != nil {
		return nil, err
	}

	// Try to bind
	list, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	if tlsConf != nil {
		list = tls.NewListener(list, tlsConf)
	}

	// Create a cmux
	mux := cmux.New(list)

	// Match connections in order:
	// First grpc, otherwise TCP
	grpcL := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	tcpL := mux.Match(cmux.Any()) // Any means anything that is not yet matched.

	// Create the protocol Server
	var grpcS *grpc.Server
	grpcS = grpc.NewServer()
	grpcServiceRegister(grpcS)

	// Use the muxed listeners for your servers
	go grpcS.Serve(grpcL)

	go func() {
		if err := mux.Serve(); !strings.Contains(err.Error(), "use of closed network connection") {
			panic(err)
		}
	}()

	stream := &CMuxTCPStreamLayer{
		advertise:  advertiseAddr,
		tlsConfig:  tlsC,
		mux:        mux,
		listener:   tcpL,
		grpcServer: grpcS,
	}

	// Verify that we have a usable advertise address
	addr, ok := stream.Addr().(*net.TCPAddr)
	if !ok {
		stream.Close()
		return nil, errNotTCP
	}
	if addr.IP.IsUnspecified() {
		stream.Close()
		return nil, errNotAdvertisable
	}

	return stream, nil
}

// Dial implements the StreamLayer interface.
func (l *CMuxTCPStreamLayer) Dial(address raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	var err error
	var conn net.Conn
	conf, err := l.tlsConfig.OutgoingTLSConfig()
	if err != nil {
		return nil, err
	}

	if conf != nil {
		conn, err = tls.DialWithDialer(dialer, "tcp", string(address), conf)
		if err != nil {
			return nil, err
		}
		conn, ok := conn.(*tls.Conn)
		if ok {
			err = conn.Handshake()
			if err != nil {
				defer conn.Close()
				return nil, fmt.Errorf("handshake failed: %v", err)
			}
		}
	} else {
		conn, err = dialer.Dial("tcp", string(address))
	}
	return conn, err
}

// Accept implements the net.Listener interface.
func (l *CMuxTCPStreamLayer) Accept() (c net.Conn, err error) {
	return l.listener.Accept()
}

// Close implements the net.Listener interface.
func (l *CMuxTCPStreamLayer) Close() (err error) {
	l.closeLock.Lock()
	defer l.closeLock.Unlock()

	if !l.closed {
		l.closed = true
	}

	if err := l.listener.Close(); err != nil {
		return err
	}
	l.grpcServer.GracefulStop()
	return nil
}

// Addr implements the net.Listener interface.
func (l *CMuxTCPStreamLayer) Addr() net.Addr {
	// Use an advertise addr if provided
	if l.advertise != nil {
		return l.advertise
	}
	return l.listener.Addr()
}

// NewCMuxTCPTransport returns a NetworkTransport that is built on top of
// a TCP streaming transport layer.
func NewCMuxTCPTransport(
	bindAddr string,
	maxPool int,
	timeout time.Duration,
	tls *tlsutil.TLSConfigurator,
	grpcServiceRegister GRPCServiceRegister,
) (*raft.NetworkTransport, error) {
	return NewCMuxTCPTransportWithLogger(bindAddr, maxPool, timeout, tls, grpcServiceRegister, log.L())
}

// NewCMuxTCPTransportWithLogger returns a NetworkTransport that is built on top of
// a TCP streaming transport layer, with log output going to the supplied Logger
func NewCMuxTCPTransportWithLogger(
	bindAddr string,
	maxPool int,
	timeout time.Duration,
	tls *tlsutil.TLSConfigurator,
	grpcServiceRegister GRPCServiceRegister,
	logger log.Logger,
) (*raft.NetworkTransport, error) {
	stream, err := NewCMuxTCPStreamLayer(bindAddr, tls, grpcServiceRegister)
	if err != nil {
		return nil, err
	}
	config := &raft.NetworkTransportConfig{
		Stream:  stream,
		MaxPool: maxPool,
		Timeout: timeout,
		Logger: logger.StdLogger(&log.StdLoggerOptions{
			InferLevels: true,
		}),
	}
	return raft.NewNetworkTransportWithConfig(config), nil
}
