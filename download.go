package main

type Download struct {
	Pieces      []Piece
	TotalLength int
	PieceLength int
	BlockSize   int
}

type Piece struct {
	Index int
	Hash  []byte
	Data  []Block
}

type Block struct {
	Index  int
	Offset int
	Length int
	Data   []byte
}

func NewDownload(pieceLength int, totalLength int, pieceHashes [][]byte) *Download {
	numPieces := (totalLength + (pieceLength) - 1) / (pieceLength)
	pieces := make([]Piece, numPieces)

	for i := range pieces {
		// Calculate number of blocks for this piece
		numBlocks := (pieceLength + BLOCK_SIZE - 1) / BLOCK_SIZE
		blocks := make([]Block, numBlocks)

		// Initialize each block
		for j := range blocks {
			blocks[j] = Block{
				Index:  j,
				Offset: j * BLOCK_SIZE,
				Length: min(BLOCK_SIZE, pieceLength-(j*BLOCK_SIZE)),
				Data:   make([]byte, 0),
			}
		}

		pieces[i] = Piece{
			Index: i,
			Hash:  pieceHashes[i],
			Data:  blocks,
		}
	}

	return &Download{
		Pieces:      pieces,
		TotalLength: totalLength,
		PieceLength: pieceLength,
		BlockSize:   BLOCK_SIZE,
	}
}

func (d *Download) GetFirstBlockOfFirstPiece() *Block {
	if len(d.Pieces) == 0 {
		return nil // No pieces available
	}
	if len(d.Pieces[0].Data) == 0 {
		return nil // No blocks available in the first piece
	}
	return &d.Pieces[0].Data[0] // Return the first block of the first piece
}
