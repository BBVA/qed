/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, n.A.
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
package member

import (
	"net"

	"github.com/bbva/qed/log"
	"github.com/hashicorp/memberlist"
)

type Type int

func (t Type) String() string {
	switch t {
	case Auditor:
		return "auditor"
	case Monitor:
		return "monitor"
	case Publisher:
		return "publisher"
	case Server:
		return "server"
	default:
		return "unknown"
	}
}

func ParseType(value string) Type {
	switch value {
	case "auditor":
		return Auditor
	case "monitor":
		return Monitor
	case "publisher":
		return Publisher
	default:
		return Server
	}
}

const (
	Auditor Type = iota
	Monitor
	Publisher
	Server
	Unknown
)

// Member is a single member of the gossip cluster.
type Peer struct {
	Name   string
	Addr   net.IP
	Port   uint16
	Meta   Meta
	Status Status
}

func (p Peer) Node() *memberlist.Node {
	return &memberlist.Node{
		Name: p.Name,
		Addr: p.Addr,
		Port: p.Port,
	}
}

func NewPeer(name, addr string, port uint16, role Type) *Peer {
	meta := Meta{
		Role: role,
	}

	return &Peer{
		Name: name,
		Addr: net.IP(addr),
		Port: port,
		Meta: meta,
	}
}

func ParsePeer(node *memberlist.Node) *Peer {
	var meta Meta
	err := meta.Decode(node.Meta)
	if err != nil {
		log.Errorf("Error parsing peer: unable to decode meta. %v", err)
	}
	return &Peer{
		Name: node.Name,
		Addr: node.Addr,
		Port: node.Port,
		Meta: meta,
	}
}
