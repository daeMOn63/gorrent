package handlers

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"

	"github.com/daeMOn63/gorrent/tracker/actions"
	"github.com/daeMOn63/gorrent/tracker/store"
)

type announce struct {
	store store.Announce
}

// NewAnnounce returns a new Handler for announce actions
func NewAnnounce(store store.Announce) actions.Handler {
	return &announce{
		store: store,
	}
}

// Handle process the announce action
func (h *announce) Handle(a actions.Action) ([]byte, error) {
	announceAction, ok := a.(*actions.Announce)
	if !ok {
		return nil, errors.New("given action is not a valid announce action")
	}

	log.Printf("announce %s %#x", actions.EventNames[announceAction.Event], announceAction.InfoHash)

	h.store.Save(announceAction)

	peers := h.store.FindPeers(announceAction.InfoHash)

	var addrs []actions.PeerAddr
	for _, p := range peers {
		if p.ID != announceAction.Peer.ID {
			addrs = append(addrs, p.PeerAddr)
		}
	}

	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, addrs); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
