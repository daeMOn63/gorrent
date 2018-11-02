package gorrent

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

// PeerAddr describes a peer address (ip and port)
type PeerAddr struct {
	IPAddr uint32
	Port   uint16
}

// Bytes return the byte representation of the PeerAddr
func (p PeerAddr) Bytes() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := binary.Write(buf, binary.BigEndian, p)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// String retuns a string representation of the PeerAddr
func (p PeerAddr) String() string {
	return fmt.Sprintf("%s:%d", int2ip(p.IPAddr), p.Port)
}

// Peer defines the peer id, and exposed ip and port
type Peer struct {
	PeerAddr
	ID PeerID
}

// NewPeer creates a new Peer
func NewPeer(id string, ip net.IP, port uint16) *Peer {
	peerID := &PeerID{}
	peerID.SetString(id)
	ipInt := ip2int(ip)

	return &Peer{
		ID: *peerID,
		PeerAddr: PeerAddr{
			IPAddr: ipInt,
			Port:   port,
		},
	}
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

// PeerID defines a custom type to store the Peer IDentifier
type PeerID [20]byte

// SetString set given string as PeerID value
// It will panic if the string is longer than the maximum size of a PeerID
func (p *PeerID) SetString(id string) {
	if len(p) < len([]byte(id)) {
		panic(fmt.Sprintf("string %s too long for a PeerID (max %d)", id, len(p)))
	}

	copy(p[:], id)
}
