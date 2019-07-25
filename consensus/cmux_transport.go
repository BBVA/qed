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
	"errors"
	io "io"
	"log"
	"net"
	"time"

	"github.com/hashicorp/raft"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

var (
	errNotAdvertisable = errors.New("local bind address is not advertisable")
	errNotTCP          = errors.New("local address is not a TCP address")
)

// TCPStreamLayer implements StreamLayer interface for plain TCP.
type TCPStreamLayer struct {
	advertise  net.Addr
	mux        cmux.CMux
	grpcServer *grpc.Server
	listener   *net.TCPListener
}

// NewCMuxTCPTransport returns a NetworkTransport that is built on top of
// a TCP streaming transport layer.
func NewCMuxTCPTransport(
	bindAddr string,
	advertise net.Addr,
	maxPool int,
	timeout time.Duration,
	logOutput io.Writer,
) (*raft.NetworkTransport, error) {
	return newTCPTransport(bindAddr, advertise, func(stream raft.StreamLayer) *raft.NetworkTransport {
		return raft.NewNetworkTransport(stream, maxPool, timeout, logOutput)
	})
}

// NewCMuxTCPTransportWithLogger returns a NetworkTransport that is built on top of
// a TCP streaming transport layer, with log output going to the supplied Logger
func NewCMuxTCPTransportWithLogger(
	bindAddr string,
	advertise net.Addr,
	maxPool int,
	timeout time.Duration,
	logger *log.Logger,
) (*raft.NetworkTransport, error) {
	return newTCPTransport(bindAddr, advertise, func(stream raft.StreamLayer) *raft.NetworkTransport {
		return raft.NewNetworkTransportWithLogger(stream, maxPool, timeout, logger)
	})
}

// NewCMuxTCPTransportWithConfig returns a NetworkTransport that is built on top of
// a TCP streaming transport layer, using the given config struct.
func NewCMuxTCPTransportWithConfig(
	bindAddr string,
	advertise net.Addr,
	config *raft.NetworkTransportConfig,
) (*raft.NetworkTransport, error) {
	return newTCPTransport(bindAddr, advertise, func(stream raft.StreamLayer) *raft.NetworkTransport {
		config.Stream = stream
		return raft.NewNetworkTransportWithConfig(config)
	})
}

func newTCPTransport(bindAddr string,
	advertise net.Addr,
	transportCreator func(stream raft.StreamLayer) *raft.NetworkTransport) (*raft.NetworkTransport, error) {
	// Try to bind
	list, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	// Create a cmux
	mux := cmux.New(list)

	// Match connections in order:
	// First grpc, otherwise TCP
	grpcL := mux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	tcpL := mux.Match(cmux.Any()) // Any means anything that is not yet matched.

	// Create the protocol Server
	grpcS := grpc.NewServer()

	// Use the muxed listeners for your servers
	go grpcS.Serve(grpcL)

	// Create stream
	stream := &TCPStreamLayer{
		advertise:  advertise,
		mux:        mux,
		grpcServer: grpcS,
		listener:   tcpL.(*net.TCPListener),
	}

	// Verify that we have a usable advertise address
	addr, ok := stream.Addr().(*net.TCPAddr)
	if !ok {
		list.Close()
		return nil, errNotTCP
	}
	if addr.IP.IsUnspecified() {
		list.Close()
		return nil, errNotAdvertisable
	}

	// Create the network transport
	trans := transportCreator(stream)
	return trans, nil
}

// Dial implements the StreamLayer interface.
func (t *TCPStreamLayer) Dial(address raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", string(address), timeout)
}

// Accept implements the net.Listener interface.
func (t *TCPStreamLayer) Accept() (c net.Conn, err error) {
	return t.listener.Accept()
}

// Close implements the net.Listener interface.
func (t *TCPStreamLayer) Close() (err error) {
	t.grpcServer.GracefulStop()
	if err := t.listener.Close(); err != nil {
		return err
	}
	return nil
}

// Addr implements the net.Listener interface.
func (t *TCPStreamLayer) Addr() net.Addr {
	// Use an advertise addr if provided
	if t.advertise != nil {
		return t.advertise
	}
	return t.listener.Addr()
}
