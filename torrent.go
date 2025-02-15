package main

type Torrent struct {
	Announce     string
	CreationDate int64
	Comment      string
	CreatedBy    string
	Info         TorrentInfo
}

type TorrentInfo struct {
	Name        string
	PieceLength int
	Pieces      string
	Length      int
	Files       []File
}

type File struct {
	Length int
	Path   []string
}
