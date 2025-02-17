package main

import (
	"log"
)

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

		// log.Println(reflect.TypeOf(info["files"]))

		if files, ok := info["files"].([]interface{}); ok {
			if len(files) > 0 {
				var tempFileStorage []File
				for _, file := range files {
					// log.Println(reflect.TypeOf(file))

					fileMap, ok := file.(map[string]interface{})
					if !ok {
						log.Println("Skipping file: not a map")
						continue
					}

					f := File{}
					if value, ok := fileMap["length"].(int); ok {
						f.Length = value
					}
					if paths, ok := fileMap["path"].([]interface{}); ok {
						var tempPathsStorage []string
						for _, path := range paths {
							if value, ok := path.([]byte); ok {
								tempPathsStorage = append(tempPathsStorage, string(value))
							}
						}
						f.Path = tempPathsStorage
					}

					tempFileStorage = append(tempFileStorage, f)
				}
				torrent.Info.Files = tempFileStorage
			}
		}

	}

	return torrent

}
