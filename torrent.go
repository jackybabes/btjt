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
	debugInterface map[string]interface{}
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

// func loadTorrentIntoStruct(data map[string]interface{}) TorrentData {
// 	torrent := TorrentData{}

// 	if value, ok := data["announce"].([]byte); ok {
// 		torrent.Announce = string(value)
// 	}

// 	if value, ok := data["creation date"].(int); ok {
// 		torrent.CreationDate = value
// 	}

// 	if info, ok := data["info"].(map[string]interface{}); ok {

// 		if value, ok := info["length"].(int); ok {
// 			torrent.Info.Length = value
// 		}

// 		if value, ok := info["name"].([]byte); ok {
// 			torrent.Info.Name = string(value)
// 		}

// 		if value, ok := info["piece length"].(int); ok {
// 			torrent.Info.PieceLength = value
// 		}

// 		if value, ok := info["pieces"].([]byte); ok {
// 			torrent.Info.Pieces = value
// 		}

// 		if value, ok := info["private"].(int); ok {
// 			torrent.Info.Private = value
// 		}

// 		// log.Println(reflect.TypeOf(info["files"]))

// 		if files, ok := info["files"].([]interface{}); ok {
// 			if len(files) > 0 {
// 				var tempFileStorage []File
// 				for _, file := range files {
// 					// log.Println(reflect.TypeOf(file))

// 					fileMap, ok := file.(map[string]interface{})
// 					if !ok {
// 						log.Println("Skipping file: not a map")
// 						continue
// 					}

// 					f := File{}
// 					if value, ok := fileMap["length"].(int); ok {
// 						f.Length = value
// 					}
// 					if paths, ok := fileMap["path"].([]interface{}); ok {
// 						var tempPathsStorage []string
// 						for _, path := range paths {
// 							if value, ok := path.([]byte); ok {
// 								tempPathsStorage = append(tempPathsStorage, string(value))
// 							}
// 						}
// 						f.Path = tempPathsStorage
// 					}

// 					tempFileStorage = append(tempFileStorage, f)
// 				}
// 				torrent.Info.Files = tempFileStorage
// 			}
// 		}

// 	}

// 	return torrent

// }
