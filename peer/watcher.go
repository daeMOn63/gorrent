package peer

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/daeMOn63/gorrent/buffer"
	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/gorrent"
	"github.com/daeMOn63/gorrent/tracker"
)

var (
	// ErrIntegrityCheckFailed is returned when the watcher fail to validate a file integrity
	ErrIntegrityCheckFailed = errors.New("integrity check failed")

	// ErrNoMoreChunk is returned when no more chunk are available to download
	ErrNoMoreChunk = errors.New("no more chunk")
)

// Watcher defines a gorrent watcher, responsible of checking the status and integrity of stored gorrent
type Watcher interface {
	Watch() error
}

type watcher struct {
	store      GorrentStore
	fs         fs.FileSystem
	fileBuffer buffer.File
	tracker    tracker.Client
	peerClient Client
}

var _ Watcher = &watcher{}

// NewWatcher creates a new gorrent watcher
func NewWatcher(store GorrentStore, fs fs.FileSystem, fileBuffer buffer.File, tracker tracker.Client, peerClient Client) Watcher {
	return &watcher{
		store:      store,
		fs:         fs,
		fileBuffer: fileBuffer,
		tracker:    tracker,
		peerClient: peerClient,
	}
}

// Watch periodically check each stored gorrent for integrity and completion
func (w *watcher) Watch() error {

	var processedEntries []gorrent.Sha1Hash

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			entries, err := w.store.All()
			if err != nil {
				log.Println(err)
			}

			for _, entry := range entries {
				found := false
				for _, pentry := range processedEntries {
					if entry.Gorrent.InfoHash() == pentry {
						found = true
					}
				}

				if found {
					continue
				}
				processedEntries = append(processedEntries, entry.Gorrent.InfoHash())

				currentHash := entry.Gorrent.InfoHash()
				log.Printf("Starting processing gorrent %s (%s)", entry.Name, currentHash.HexString())

				go func() {
					for {
						// Refresh entry from db
						currentEntry, err := w.store.Get(currentHash)
						if err != nil {
							log.Println(err)
						}

						log.Printf("Entry %s (%s) status: %s", currentEntry.Name, currentEntry.Gorrent.InfoHash().HexString(), currentEntry.Status)
						switch currentEntry.Status {
						case StatusNew:
							err = w.processNew(currentEntry)
						case StatusReady:
							err = w.processReady(currentEntry)
						case StatusCheck:
							err = w.processCheck(currentEntry)
						case StatusDownloading:
							err = w.processDownloading(currentEntry)
						case StatusCompleted:
							w.fs.Remove(currentEntry.TmpFileName())
							log.Printf("Gorrent %s (%s) is complete.", currentEntry.Name, currentEntry.Gorrent.InfoHash().HexString())
							return
						}

						if err != nil {
							log.Println(err)
						}
					}
				}()
			}
		}
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}

func (w *watcher) processCheck(entry *GorrentEntry) error {
	log.Printf("Checking %s (%s)", entry.Name, entry.Gorrent.InfoHash().HexString())
	chunkedFile, err := w.fileBuffer.Open(entry.TmpFileName(), entry.Gorrent.PieceLength)
	if err != nil {
		return err
	}
	defer chunkedFile.Close()

	var currentOffset int64
	for _, file := range entry.Gorrent.Files {
		data, err := chunkedFile.Read(file.Length, currentOffset)
		if err != nil {
			return err
		}

		fileHash := sha1.Sum(data)

		log.Printf("Checking integrity for file %s", file.Name)
		if fileHash != file.Hash {
			log.Printf("Integrity check failed %v != %v", fileHash, file.Hash)
			return ErrIntegrityCheckFailed
		}

		filePath := filepath.Join(entry.Path, file.Name)
		if err := w.fs.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		outFile, err := w.fs.Create(filePath)
		if err != nil {
			return err
		}

		n, err := outFile.Write(data)
		if err != nil {
			return err
		}
		log.Printf("Wrote file %s (%d bytes)", filePath, n)
		outFile.Close()

		currentOffset += file.Length
	}

	entry.Status = StatusCompleted
	return w.store.Save(entry)
}

func getMissingChunkID(entry *GorrentEntry) (int64, error) {
	completedChunks := entry.CompletedChunks
	if len(completedChunks) == 0 {
		return 0, nil
	}

	if len(completedChunks) == len(entry.Gorrent.Pieces) {
		return 0, ErrNoMoreChunk
	}

	return completedChunks[len(completedChunks)-1] + 1, nil
}

func (w *watcher) processDownloading(entry *GorrentEntry) error {
	if len(entry.PeerAddrs) <= 0 {
		time.Sleep(500 * time.Millisecond)
		return fmt.Errorf("No peers to download %s (%s)", entry.Name, entry.Gorrent.InfoHash().HexString())
	}

	chunkedFile, err := w.fileBuffer.Open(entry.TmpFileName(), entry.Gorrent.PieceLength)
	if err != nil {
		return err
	}
	defer chunkedFile.Close()

	chunkID, err := getMissingChunkID(entry)
	if err == ErrNoMoreChunk {
		log.Printf("No more chunk to download for %s (%s)", entry.Name, entry.Gorrent.InfoHash().HexString())
		entry.Status = StatusCheck

		return w.store.Save(entry)
	}

	if err != nil {
		return err
	}

	log.Printf("Downloading chunk %d from %s (%s)", chunkID, entry.Name, entry.Gorrent.InfoHash().HexString())

	for _, peerAddr := range entry.PeerAddrs {

		request := &ChunkRequest{
			InfoHash: entry.Gorrent.InfoHash(),
			ChunkID:  chunkID,
		}

		log.Printf("Contacting peer %s", peerAddr.String())
		data, err := w.peerClient.GetPiece(peerAddr, request, entry.Gorrent.PieceLength)
		if err != nil {
			log.Printf("Peer:GetPiece (%v) error: %s", peerAddr, err)

			continue
		}

		hash := sha1.Sum(data)
		if hash != entry.Gorrent.Pieces[chunkID] {
			log.Printf("Integrity check failed for chunk %d, expected %v, got %v", chunkID, entry.Gorrent.Pieces[chunkID], hash)

			continue
		}

		if err := chunkedFile.WriteChunk(chunkID, data); err != nil {
			log.Printf("Peer:WriteChunk (%v) error: %s", peerAddr, err)

			continue
		}

		entry.Downloaded += uint64(len(data))
		entry.CompletedChunks = append(entry.CompletedChunks, chunkID)

		if err := w.store.Save(entry); err != nil {
			log.Printf("Store:Save error: %s", err)

			continue
		}

		log.Printf("Completed downloading chunk %d", chunkID)
	}

	return nil
}

// processReady creates the temp file buffer for storing downloaded data
func (w *watcher) processReady(entry *GorrentEntry) error {
	err := w.fileBuffer.Create(entry.TmpFileName(), int64(entry.Gorrent.TotalFileSize()))
	if err != nil {
		return err
	}

	entry.Status = StatusDownloading

	return w.store.Save(entry)
}

// processNew check if the gorrent is alreay completed and pass integrity checks or set its status to ready
func (w *watcher) processNew(entry *GorrentEntry) error {
	err := w.fs.MkdirAll(entry.Path, 0755)
	if err != nil {
		return err
	}

	isCompleted := true
	for _, gorrentFile := range entry.Gorrent.Files {
		filePath := filepath.Join(entry.Path, gorrentFile.Name)
		f, err := w.fs.Open(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("gorrent %s (%s) is not complete yet.", entry.Name, entry.Gorrent.InfoHash().HexString())
				isCompleted = false
				break
			}

			return err
		}

		log.Printf("checking integrity for file %s\n", filePath)
		if err := checkIntegrity(f, gorrentFile.Hash); err != nil {
			log.Printf("integrity check failed for file %s\n", filePath)

			entry.Status = StatusCorrupted
			w.store.Save(entry)

			return err
		}

		f.Close()
	}

	if isCompleted {
		entry.Downloaded = entry.Gorrent.TotalFileSize()
		entry.Status = StatusCompleted
	} else {
		log.Printf("gorrent %s (%s) is ready for download\n", entry.Name, entry.Gorrent.InfoHash().HexString())
		entry.Status = StatusReady
	}

	return w.store.Save(entry)
}

func checkIntegrity(r io.Reader, expectedHash gorrent.Sha1Hash) error {
	hash := sha1.New()
	_, err := io.Copy(hash, r)
	if err != nil {
		return err
	}

	var sha1Hash gorrent.Sha1Hash
	copy(sha1Hash[:], hash.Sum(nil))

	if sha1Hash.HexString() != expectedHash.HexString() {
		return ErrIntegrityCheckFailed
	}

	return nil
}
