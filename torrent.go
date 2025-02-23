package main

import (
	"crypto/sha1"
	"log"
	"os"
)

type Torrent struct {
	Data           TorrentData
	BencodedData   []byte
	FilePath       string
	InfoHash       [20]byte
	debugInterface map[string]any
	multiFile      bool
	downloadLength int
	numberOfPieces int
	PieceHashes    [][]byte
	Trackers       []*Tracker
	Uploaded       int
	Downloaded     int
	Left           int
}

type TorrentData struct {
	Announce       string
	Announce__List [][]string
	Creation_Date  int
	Comment        string
	CreatedBy      string
	Info           TorrentInfo
}

type TorrentInfo struct {
	Name         string
	Piece_Length int
	Pieces       []byte
	Length       int
	Files        []File
	Private      int
}

type File struct {
	Length int
	Path   []string
}

func NewTorrent(fp string) Torrent {
	t := Torrent{FilePath: fp}
	t.readBytesFromFile()

	// Check file is dict
	if t.BencodedData[0] != 'd' {
		log.Panicln("Torrent File not dict")
	}

	t.unMarshalBencodedData()
	t.calcLength()
	t.pieceHashes()

	return t
}

func (t *Torrent) readBytesFromFile() {
	data, err := os.ReadFile(t.FilePath)
	if err != nil {
		log.Fatalln(err)
	}
	t.BencodedData = data
}

func (t *Torrent) unMarshalBencodedData() {

	key := "info"
	interfaceMap, infoHashBytes := bencode_decode_byte_data_of_key(t.BencodedData, key)

	if len(infoHashBytes) == 0 {
		log.Panicln("Could not find info section")
	}
	t.InfoHash = sha1.Sum(infoHashBytes)

	populateStruct(&t.Data, interfaceMap)
	t.debugInterface = interfaceMap
}

func (t *Torrent) calcLength() {

	// Check multi File
	if len(t.Data.Info.Files) > 0 {
		t.multiFile = true
		for _, file := range t.Data.Info.Files {
			t.downloadLength += file.Length
		}
	} else {
		t.multiFile = false
		t.downloadLength = t.Data.Info.Length
	}
	t.numberOfPieces = (t.downloadLength / t.Data.Info.Piece_Length) + 1

	t.Left = t.downloadLength

}

func (t *Torrent) pieceHashes() {
	// // Create slice of hashes
	pieceBuffer := t.Data.Info.Pieces
	for range t.numberOfPieces {
		t.PieceHashes = append(t.PieceHashes, pieceBuffer[:20])
		pieceBuffer = pieceBuffer[20:]
	}
}

func (t *Torrent) createTrackers() {
	// does the annouce list contain the top level tracker always?
	if len(t.Data.Announce__List) == 0 {
		if t.Data.Announce == "" {
			log.Panicln("No trackers")
		}
		t.Trackers = append(t.Trackers, NewTracker(t.Data.Announce, 0))
	} else {
		for tier, trackerList := range t.Data.Announce__List {
			for _, tracker := range trackerList {
				t.Trackers = append(t.Trackers, NewTracker(tracker, tier))
			}
		}
	}
}
