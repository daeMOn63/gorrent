package actions

import (
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

// AnnounceStatus holds information about the current download state
type AnnounceStatus struct {
	Downloaded uint64
	Uploaded   uint64
}

// Announce holds announce action data
type Announce struct {
	InfoHash gorrent.Sha1Hash
	Peer     gorrent.Peer
	Status   AnnounceStatus
	Event    AnnounceEvent
}

// ID contains the action identifier
func (a *Announce) ID() ID {
	return AnnounceID
}
