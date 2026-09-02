package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"sync"
	"time"
)

// How long to wait for a requested block before another peer may claim it.
const requestTimeout = 15 * time.Second

type Download struct {
	Pieces      []Piece
	TotalLength int
	PieceLength int
	BlockSize   int
	Storage     *Storage

	mu             sync.Mutex
	completedCount int
	Done           chan struct{}
	doneClosed     bool
}

type Piece struct {
	Index  int
	Hash   []byte
	Length int
	Data   []Block
	done   bool
}

type Block struct {
	Index       int
	Offset      int
	Length      int
	Data        []byte
	Downloading bool
	RequestedAt time.Time
}

func NewDownload(pieceLength int, totalLength int, pieceHashes [][]byte) *Download {
	numPieces := (totalLength + pieceLength - 1) / pieceLength
	pieces := make([]Piece, numPieces)

	for i := range pieces {
		thisPieceLen := pieceLength
		if i == numPieces-1 {
			if rem := totalLength - (numPieces-1)*pieceLength; rem > 0 {
				thisPieceLen = rem
			}
		}

		numBlocks := (thisPieceLen + BLOCK_SIZE - 1) / BLOCK_SIZE
		blocks := make([]Block, numBlocks)
		for j := range blocks {
			blocks[j] = Block{
				Index:  j,
				Offset: j * BLOCK_SIZE,
				Length: min(BLOCK_SIZE, thisPieceLen-(j*BLOCK_SIZE)),
			}
		}

		pieces[i] = Piece{
			Index:  i,
			Hash:   pieceHashes[i],
			Length: thisPieceLen,
			Data:   blocks,
		}
	}

	return &Download{
		Pieces:      pieces,
		TotalLength: totalLength,
		PieceLength: pieceLength,
		BlockSize:   BLOCK_SIZE,
		Done:        make(chan struct{}),
	}
}

// NextRequest picks the next block that peer p is able to provide. It returns
// (-1, nil) when there is nothing this peer can usefully download right now.
func (d *Download) NextRequest(p *Peer) (pieceIndex int, block *Block) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for pi := range d.Pieces {
		piece := &d.Pieces[pi]
		if piece.done || !p.HasPiece(pi) {
			continue
		}
		for bi := range piece.Data {
			b := &piece.Data[bi]
			if len(b.Data) > 0 {
				continue // already have it
			}
			if b.Downloading && now.Sub(b.RequestedAt) < requestTimeout {
				continue // in flight elsewhere
			}
			b.Downloading = true
			b.RequestedAt = now
			return pi, b
		}
	}
	return -1, nil
}

// SaveBlock stores a received block. When its piece becomes complete the piece
// is hash-verified and flushed to storage. Returns true once the whole download
// is finished.
func (d *Download) SaveBlock(pieceIndex, begin int, data []byte) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if pieceIndex < 0 || pieceIndex >= len(d.Pieces) {
		return false, fmt.Errorf("invalid piece index %d", pieceIndex)
	}
	piece := &d.Pieces[pieceIndex]
	if piece.done {
		return d.doneClosed, nil
	}

	blockIndex := begin / BLOCK_SIZE
	if blockIndex < 0 || blockIndex >= len(piece.Data) {
		return false, fmt.Errorf("invalid block offset %d for piece %d", begin, pieceIndex)
	}
	b := &piece.Data[blockIndex]
	if b.Offset != begin {
		return false, fmt.Errorf("misaligned block offset %d for piece %d", begin, pieceIndex)
	}
	if len(b.Data) > 0 {
		return false, nil // duplicate, ignore
	}

	b.Data = append([]byte(nil), data...)
	b.Downloading = false

	for i := range piece.Data {
		if len(piece.Data[i].Data) == 0 {
			return false, nil // piece not complete yet
		}
	}

	// Piece complete: assemble, verify, persist.
	assembled := make([]byte, 0, piece.Length)
	for i := range piece.Data {
		assembled = append(assembled, piece.Data[i].Data...)
	}
	hash := sha1.Sum(assembled)
	if !bytes.Equal(hash[:], piece.Hash) {
		log.Printf("piece %d failed hash check, will retry", pieceIndex)
		for i := range piece.Data {
			piece.Data[i].Data = nil
			piece.Data[i].Downloading = false
		}
		return false, nil
	}

	if d.Storage != nil {
		if err := d.Storage.WritePiece(pieceIndex, d.PieceLength, assembled); err != nil {
			return false, fmt.Errorf("write piece %d: %w", pieceIndex, err)
		}
	}
	for i := range piece.Data {
		piece.Data[i].Data = nil // release memory
	}
	piece.done = true
	d.completedCount++
	log.Printf("piece %d complete (%d/%d)", pieceIndex, d.completedCount, len(d.Pieces))

	if d.completedCount == len(d.Pieces) && !d.doneClosed {
		d.doneClosed = true
		close(d.Done)
	}
	return d.doneClosed, nil
}

func (d *Download) IsComplete() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.doneClosed
}

func (d *Download) Progress() (completed, total int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.completedCount, len(d.Pieces)
}
