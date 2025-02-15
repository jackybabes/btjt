package main

type Torrent struct {
	Announce     string
	CreationDate int
	Comment      string
	CreatedBy    string
	Info         TorrentInfo
}

type TorrentInfo struct {
	Name        string
	PieceLength int
	Pieces      []byte
	Length      int
	Files       []File
	Private     int
}

type File struct {
	Length int
	Path   []string
}
