package store

import (
	"time"

	"github.com/daeMOn63/gorrent/gorrent"
	"github.com/daeMOn63/gorrent/tracker/actions"
)

// Announce defines a tracker storage
type Announce interface {
	Save(announce *actions.Announce)
	Find(infoHash gorrent.Sha1Hash, peerID gorrent.PeerID) *StoredAnnounce
	FindPeers(infoHash gorrent.Sha1Hash, maxAge time.Duration) []gorrent.Peer
}

// AnnounceMemory defines a tracker storage using memory only
type AnnounceMemory struct {
	announces []*StoredAnnounce
}

var _ Announce = &AnnounceMemory{}
var _ Announce = &DummyAnnounce{}

// NewAnnounceMemory creates a new in memory store
func NewAnnounceMemory() *AnnounceMemory {
	return &AnnounceMemory{}
}

// StoredAnnounce is how actions.Announce get stored
type StoredAnnounce struct {
	Announce    *actions.Announce
	LastUpdated time.Time
}

// Save add an announce action to the memory store
func (m *AnnounceMemory) Save(announce *actions.Announce) {
	var sa *StoredAnnounce
	sa = m.Find(announce.InfoHash, announce.Peer.ID)
	if sa == nil {
		sa = &StoredAnnounce{}
		m.announces = append(m.announces, sa)
	}

	sa.Announce = announce
	sa.LastUpdated = time.Now()
}

// Find retrieve a stored announce
func (m *AnnounceMemory) Find(infoHash gorrent.Sha1Hash, peerID gorrent.PeerID) *StoredAnnounce {
	for _, a := range m.announces {
		if a.Announce.InfoHash == infoHash && a.Announce.Peer.ID == peerID {
			return a
		}
	}

	return nil
}

// FindPeers retrieve all peers on a given infoHash
func (m *AnnounceMemory) FindPeers(infoHash gorrent.Sha1Hash, maxAge time.Duration) []gorrent.Peer {
	var peers []gorrent.Peer

	for _, a := range m.announces {
		if a.Announce.InfoHash == infoHash {
			limit := time.Now().Add(-maxAge)
			if a.LastUpdated.After(limit) {
				peers = append(peers, a.Announce.Peer)
			}
		}
	}

	return peers
}

// DummyAnnounce provides a configurable Announce store
type DummyAnnounce struct {
	SaveFunc      func(announce *actions.Announce)
	FindFunc      func(infoHash gorrent.Sha1Hash, peerID gorrent.PeerID) *StoredAnnounce
	FindPeersFunc func(infoHash gorrent.Sha1Hash, maxAge time.Duration) []gorrent.Peer
}

// Save calls SaveFunc
func (d *DummyAnnounce) Save(announce *actions.Announce) {
	d.SaveFunc(announce)
}

// Find calls FindFunc
func (d *DummyAnnounce) Find(infoHash gorrent.Sha1Hash, peerID gorrent.PeerID) *StoredAnnounce {
	return d.FindFunc(infoHash, peerID)
}

// FindPeers calls FindPeersFunc
func (d *DummyAnnounce) FindPeers(infoHash gorrent.Sha1Hash, maxAge time.Duration) []gorrent.Peer {
	return d.FindPeersFunc(infoHash, maxAge)
}
