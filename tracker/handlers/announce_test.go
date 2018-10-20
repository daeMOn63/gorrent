package handlers

import (
	"reflect"
	"testing"

	"github.com/daeMOn63/gorrent/gorrent"

	"github.com/daeMOn63/gorrent/tracker/actions"
	"github.com/daeMOn63/gorrent/tracker/store"
)

func TestAnnounce(t *testing.T) {
	t.Run("Handle fail on invalid action", func(t *testing.T) {
		store := &store.DummyAnnounce{}
		h := NewAnnounce(store)

		action := &actions.DummyAction{}
		out, err := h.Handle(action)
		if out != nil {
			t.Fatalf("Expected out to be nil, got %v", out)
		}

		if err != ErrBadAction {
			t.Fatalf("Expected err to be %v, got %v", ErrBadAction, err)
		}
	})

	t.Run("Handle saves action in store and return peers address for given InfoHash", func(t *testing.T) {
		expectedInfoHash := gorrent.RandomSha1Hash()

		expectedAction := &actions.Announce{
			InfoHash: expectedInfoHash,
			Event:    actions.AnnounceEventStarted,
			Peer: actions.Peer{
				ID: actions.PeerID(gorrent.RandomSha1Hash()),
				PeerAddr: actions.PeerAddr{
					IPAddr: 1,
					Port:   2,
				},
			},
		}

		expectedPeer1 := actions.Peer{
			ID: actions.PeerID(gorrent.RandomSha1Hash()),
			PeerAddr: actions.PeerAddr{
				IPAddr: 5,
				Port:   6,
			},
		}

		expectedPeer2 := actions.Peer{
			ID: actions.PeerID(gorrent.RandomSha1Hash()),
			PeerAddr: actions.PeerAddr{
				IPAddr: 7,
				Port:   8,
			},
		}

		expectedPeers := []actions.Peer{expectedPeer1, expectedPeer2}

		store := &store.DummyAnnounce{
			SaveFunc: func(announce *actions.Announce) {
				if reflect.DeepEqual(announce, expectedAction) == false {
					t.Fatalf("Expected action to be %#v, got %#v", expectedAction, announce)
				}
			},
			FindPeersFunc: func(infoHash gorrent.Sha1Hash) []actions.Peer {
				if reflect.DeepEqual(infoHash, expectedInfoHash) == false {
					t.Fatalf("Expected infoHash to be %v, got %v", expectedInfoHash, infoHash)
				}

				return expectedPeers
			},
		}

		h := NewAnnounce(store)

		out, err := h.Handle(expectedAction)
		if err != nil {
			t.Fatalf("Expected no error, got %s", err)
		}

		addr1, err := expectedPeer1.PeerAddr.Bytes()
		if err != nil {
			t.Fatal(err)
		}
		addr2, err := expectedPeer2.PeerAddr.Bytes()
		if err != nil {
			t.Fatal(err)
		}

		expectedOut := addr1
		expectedOut = append(expectedOut, addr2...)

		if reflect.DeepEqual(out, expectedOut) == false {
			t.Fatalf("Expected out to be %v, got %v", expectedOut, out)
		}
	})
}
