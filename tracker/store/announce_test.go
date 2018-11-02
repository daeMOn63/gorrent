package store

import (
	"reflect"
	"testing"
	"time"

	"github.com/daeMOn63/gorrent/gorrent"
	"github.com/daeMOn63/gorrent/tracker/actions"
)

func TestAnnounceMemorySave(t *testing.T) {
	t.Run("Save append to the slice", func(t *testing.T) {
		s := NewAnnounceMemory()

		a := &actions.Announce{
			Event:    actions.AnnounceEventStarted,
			InfoHash: gorrent.RandomSha1Hash(),
			Peer: gorrent.Peer{
				ID: gorrent.PeerID(gorrent.RandomSha1Hash()),
			},
		}

		if len(s.announces) != 0 {
			t.Fatalf("Expected internal announces len to be 0, got %d", len(s.announces))
		}

		beforeSave := time.Now()
		s.Save(a)

		if len(s.announces) != 1 {
			t.Fatalf("Expected internal announces len to be 1, got %d", len(s.announces))
		}

		storedAnnounce := s.announces[0]

		if storedAnnounce.LastUpdated.Before(beforeSave) {
			t.Fatalf("Expected storedAnnounce lastUpdated to be before %s, got %s", beforeSave, storedAnnounce.LastUpdated)
		}

		if reflect.DeepEqual(storedAnnounce.Announce, a) == false {
			t.Fatalf("Expected announce to be %#v, got %#v", a, storedAnnounce.Announce)
		}

		a2 := &actions.Announce{
			Event:    actions.AnnounceEventCompleted,
			InfoHash: gorrent.RandomSha1Hash(),
			Peer: gorrent.Peer{
				ID: gorrent.PeerID(gorrent.RandomSha1Hash()),
			},
		}

		s.Save(a2)

		if len(s.announces) != 2 {
			t.Fatalf("Expected internal announces len to be 2, got %d", len(s.announces))
		}

		if reflect.DeepEqual(s.announces[1].Announce, a2) == false {
			t.Fatalf("Expected announce to be %#v, got %#v", a2, s.announces[1].Announce)
		}
	})

	t.Run("Save twice same announce update it", func(t *testing.T) {
		s := NewAnnounceMemory()

		a := &actions.Announce{
			Event:    actions.AnnounceEventCompleted,
			InfoHash: gorrent.RandomSha1Hash(),
			Peer: gorrent.Peer{
				ID: gorrent.PeerID(gorrent.RandomSha1Hash()),
			},
		}

		s.Save(a)

		initialTime := s.announces[0].LastUpdated

		s.Save(a)

		if len(s.announces) != 1 {
			t.Fatalf("Expected internal announces len to be 1, got %d", len(s.announces))
		}

		if !s.announces[0].LastUpdated.After(initialTime) {
			t.Fatalf("Expected lastUpdate to be after %s, got %s", initialTime, s.announces[0].LastUpdated)
		}
	})
}

func TestAnnounceMemoryFindPeers(t *testing.T) {
	t.Run("FindPeers returns proper peers", func(t *testing.T) {
		s := NewAnnounceMemory()

		infoHash1 := gorrent.RandomSha1Hash()
		infoHash2 := gorrent.RandomSha1Hash()

		expectedPeers1 := []gorrent.Peer{
			{ID: gorrent.PeerID(gorrent.RandomSha1Hash())},
			{ID: gorrent.PeerID(gorrent.RandomSha1Hash())},
		}
		expectedPeers2 := []gorrent.Peer{
			{ID: gorrent.PeerID(gorrent.RandomSha1Hash())},
		}

		a1 := &actions.Announce{
			Event:    actions.AnnounceEventCompleted,
			InfoHash: infoHash1,
			Peer:     expectedPeers1[0],
		}

		a2 := &actions.Announce{
			Event:    actions.AnnounceEventCompleted,
			InfoHash: infoHash1,
			Peer:     expectedPeers1[1],
		}

		a3 := &actions.Announce{
			Event:    actions.AnnounceEventCompleted,
			InfoHash: infoHash2,
			Peer:     expectedPeers2[0],
		}

		s.Save(a1)
		s.Save(a2)
		s.Save(a3)

		peers := s.FindPeers(infoHash1, 1*time.Second)
		if reflect.DeepEqual(peers, expectedPeers1) == false {
			t.Fatalf("Expected peers to be %#v, got %#v", expectedPeers1, peers)
		}

		peers = s.FindPeers(infoHash2, 1*time.Second)
		if reflect.DeepEqual(peers, expectedPeers2) == false {
			t.Fatalf("Expected peers to be %#v, got %#v", expectedPeers2, peers)
		}
	})

	t.Run("FindPeers returns no peer by default", func(t *testing.T) {
		s := NewAnnounceMemory()

		peers := s.FindPeers(gorrent.RandomSha1Hash(), 1*time.Second)
		if len(peers) != 0 {
			t.Fatalf("Expected peers len to be 0, got %d", len(peers))
		}
	})
}

func TestDummyAnnounce(t *testing.T) {
	expectedAnnounce := &actions.Announce{
		Event: actions.AnnounceEventStarted,
	}

	saveCalled := false

	expectedInfoHash := gorrent.RandomSha1Hash()
	expectedPeerID := gorrent.PeerID(gorrent.RandomSha1Hash())

	expectedStoredAnnounce := &StoredAnnounce{
		Announce: &actions.Announce{
			Event: actions.AnnounceEventCompleted,
		},
		LastUpdated: time.Now(),
	}

	expectedPeers := []gorrent.Peer{
		{ID: gorrent.PeerID(gorrent.RandomSha1Hash())},
		{ID: gorrent.PeerID(gorrent.RandomSha1Hash())},
	}

	expectedMaxAge := 1 * time.Second

	d := &DummyAnnounce{
		SaveFunc: func(a *actions.Announce) {
			if reflect.DeepEqual(a, expectedAnnounce) == false {
				t.Fatalf("Expected announce to be %#v, got %#v", expectedAnnounce, a)
			}
			saveCalled = true
		},
		FindFunc: func(infoHash gorrent.Sha1Hash, peerID gorrent.PeerID) *StoredAnnounce {
			if reflect.DeepEqual(infoHash, expectedInfoHash) == false {
				t.Fatalf("Expected infoHash to be %#v, got %#v", expectedInfoHash, infoHash)
			}
			if reflect.DeepEqual(peerID, expectedPeerID) == false {
				t.Fatalf("Expected peerID to be %#v, got %#v", expectedPeerID, peerID)
			}

			return expectedStoredAnnounce
		},
		FindPeersFunc: func(infoHash gorrent.Sha1Hash, maxAge time.Duration) []gorrent.Peer {
			if reflect.DeepEqual(infoHash, expectedInfoHash) == false {
				t.Fatalf("Expected infoHash to be %#v, got %#v", expectedInfoHash, infoHash)
			}

			if maxAge != expectedMaxAge {
				t.Fatalf("Expected maxAge to be %#v, got %#v", expectedMaxAge, maxAge)
			}

			return expectedPeers
		},
	}

	d.Save(expectedAnnounce)
	if saveCalled == false {
		t.Fatalf("Expected SaveFunc to be called")
	}

	storedAnnounce := d.Find(expectedInfoHash, expectedPeerID)
	if reflect.DeepEqual(storedAnnounce, expectedStoredAnnounce) == false {
		t.Fatalf("Expected stored announce to be %#v, got %#v", expectedStoredAnnounce, storedAnnounce)
	}

	peers := d.FindPeers(expectedInfoHash, expectedMaxAge)
	if reflect.DeepEqual(peers, expectedPeers) == false {
		t.Fatalf("Expected peers to be %#v, got %#v", expectedPeers, peers)
	}
}
