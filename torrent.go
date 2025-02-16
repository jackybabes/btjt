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

func loadTorrentIntoStruct(data map[string]interface{}) Torrent {
	torrent := Torrent{}
	_ = data

	if value, ok := data["announce"].([]byte); ok {
		torrent.Announce = string(value)
	}

	if value, ok := data["creation date"].(int); ok {
		torrent.CreationDate = value
	}

	if info, ok := data["info"].(map[string]interface{}); ok {

		if value, ok := info["length"].(int); ok {
			torrent.Info.Length = value
		}

		if value, ok := info["name"].([]byte); ok {
			torrent.Info.Name = string(value)
		}

		if value, ok := info["piece length"].(int); ok {
			torrent.Info.PieceLength = value
		}

		if value, ok := info["pieces"].([]byte); ok {
			torrent.Info.Pieces = value
		}

		if value, ok := info["private"].(int); ok {
			torrent.Info.Private = value
		}

	}

	return torrent

}
