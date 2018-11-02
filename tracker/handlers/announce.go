package handlers

import (
	"errors"
	"log"
	"time"

	"github.com/daeMOn63/gorrent/tracker/actions"
	"github.com/daeMOn63/gorrent/tracker/store"
)

var (
	// ErrBadAction is returned when the handler receive an unexpected action
	ErrBadAction = errors.New("given action is not a valid announce action")
)

type announce struct {
	store      store.Announce
	maxPeerAge time.Duration
}

// NewAnnounce returns a new Handler for announce actions
func NewAnnounce(store store.Announce, maxPeerAge time.Duration) actions.Handler {
	return &announce{
		store:      store,
		maxPeerAge: maxPeerAge,
	}
}

// Handle process the announce action
func (h *announce) Handle(a actions.Action) ([]byte, error) {
	announceAction, ok := a.(*actions.Announce)
	if !ok {
		return nil, ErrBadAction
	}

	log.Printf("announce %s from %s - %#x", announceAction.Event.Name(), announceAction.Peer.ID, announceAction.InfoHash)

	h.store.Save(announceAction)

	peers := h.store.FindPeers(announceAction.InfoHash, h.maxPeerAge)

	var out []byte
	for _, p := range peers {
		if p.ID != announceAction.Peer.ID {
			b, err := p.Bytes()
			if err != nil {
				return nil, err
			}

			out = append(out, b...)
		}
	}

	return out, nil
}
