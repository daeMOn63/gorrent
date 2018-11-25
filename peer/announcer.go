package peer

import (
	"log"
	"time"

	"github.com/daeMOn63/gorrent/tracker"
	"github.com/daeMOn63/gorrent/tracker/actions"
)

// Announcer interface
type Announcer interface {
	AnnounceForever() error
}

type announcer struct {
	store    GorrentStore
	tracker  tracker.Client
	interval time.Duration
}

var _ Announcer = &announcer{}

// NewAnnouncer creates a new Announcer
func NewAnnouncer(store GorrentStore, tracker tracker.Client, interval time.Duration) Announcer {
	return &announcer{
		store:    store,
		tracker:  tracker,
		interval: interval,
	}
}

func (a *announcer) AnnounceForever() error {
	ticker := time.NewTicker(a.interval)
	log.Printf("Starting announcer")
	for range ticker.C {
		err := a.Announce()
		if err != nil {
			log.Printf("Announce error: %s", err)
		}
	}

	return nil
}

func (a *announcer) Announce() error {
	entries, err := a.store.All()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Status == StatusNew {
			log.Printf("Skipping announce for %s (%s): status new", entry.Name, entry.Gorrent.InfoHash().HexString())
			continue
		}

		log.Printf("Announcing %s (%s)", entry.Name, entry.Gorrent.InfoHash().HexString())
		peers, err := a.tracker.Announce(entry.Gorrent, actions.AnnounceEventStarted, actions.AnnounceStatus{
			Downloaded: 0,
			Uploaded:   0,
		})
		if err != nil {
			return err
		}

		entry.PeerAddrs = peers
		if err := a.store.Save(entry); err != nil {
			return err
		}
		log.Printf("Got %d peers: %s for %s (%s)", len(peers), peers, entry.Name, entry.Gorrent.InfoHash().HexString())
	}

	return nil
}
