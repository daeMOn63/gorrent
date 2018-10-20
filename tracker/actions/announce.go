package actions

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/daeMOn63/gorrent/gorrent"
)

const (
	// AnnounceEventStarted is sent when the client is starting downloading the gorrent
	AnnounceEventStarted AnnounceEvent = 0x1
	// AnnounceEventStopped is sent when the client cancel or pause the gorrent download
	AnnounceEventStopped AnnounceEvent = 0x2
	// AnnounceEventCompleted event is sent when the client finished downloading the gorrent
	AnnounceEventCompleted AnnounceEvent = 0x3
)

var (
	eventNamesMap = map[AnnounceEvent]string{
		AnnounceEventStarted:   "started",
		AnnounceEventStopped:   "stopped",
		AnnounceEventCompleted: "completed",
	}
)

// AnnounceEvent defines a custom type for announce Events
type AnnounceEvent uint8

// Name returns a human readable string naming the event
func (e AnnounceEvent) Name() string {
	return eventNamesMap[e]
}

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

// Peer defines the peer id, and exposed ip and port
type Peer struct {
	PeerAddr
	ID PeerID
}

// PeerID defines a custom type to store the Peer IDentifier
type PeerID [20]byte

// SetString set given string as PeerID value
// It will panic if the string is longer than the maximum size of a PeerID
func (p PeerID) SetString(id string) {
	if len(p) < len(id) {
		panic(fmt.Sprintf("string %s too long for a PeerID (max %d)", id, len(p)))
	}

	copy(p[:], id)
}

// AnnounceStatus holds information about the current download state
type AnnounceStatus struct {
	Downloaded uint64
	Uploaded   uint64
}

// Announce holds announce action data
type Announce struct {
	InfoHash gorrent.Sha1Hash
	Peer     Peer
	Status   AnnounceStatus
	Event    AnnounceEvent
}

// ID contains the action identifier
func (a *Announce) ID() ID {
	return AnnounceID
}
