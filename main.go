package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"io"
	"log"
	"os"
	"strconv"
)

func readTorrent(filename string) []byte {
	data, _ := os.ReadFile(filename)
	return data
}

func bytesToInt(b []byte) int {
	s := string(b)
	i, _ := strconv.Atoi(s)
	return i
}

func getBytes(reader *bufio.Reader) []byte {
	var lengthBytes []byte
	for {
		b, _ := reader.ReadByte()
		if b == ':' {
			break
		}
		lengthBytes = append(lengthBytes, b)
	}

	length := bytesToInt(lengthBytes)
	bytesBuffer := make([]byte, length)
	_, _ = reader.Read(bytesBuffer)

	// log.Printf("%v, %v", length, string(bytesBuffer))
	return bytesBuffer
}

func getInt(reader *bufio.Reader) int {
	var intBytes []byte
	for {
		b, _ := reader.ReadByte()
		if b == 'i' {
			continue
		}
		if b == 'e' {
			break
		}
		intBytes = append(intBytes, b)
	}
	i := bytesToInt(intBytes)
	return i
}

func getList(reader *bufio.Reader) []interface{} {
	var listData []interface{}

	_, _ = reader.ReadByte()
	for {
		nextBytes, _ := reader.Peek(1)
		nextByte := nextBytes[0]

		switch nextByte {
		case 'i':
			// IntType
			listData = append(listData, getInt(reader))
		case 'l':
			// ListType
			listData = append(listData, getList(reader))
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			listData = append(listData, getBytes(reader))
		case 'd':
			// DictType
			listData = append(listData, getDict(reader))
		case 'e':
			_, _ = reader.ReadByte()
			return listData
		}
	}
}

func getDict(reader *bufio.Reader) map[string]interface{} {
	dictData := make(map[string]interface{})
	_, _ = reader.ReadByte()

	for {
		// get key
		nextBytes, _ := reader.Peek(1)
		nextByte := nextBytes[0]

		if nextByte == 'e' {
			_, _ = reader.ReadByte()
			if infoSection {
				infoSectionEnd = bufferLength - reader.Buffered()
				infoSection = false
			}
			return dictData
		}

		var key string
		switch nextByte {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			key = string(getBytes(reader))
		default:
			log.Fatalln("key must be string")
		}

		if key == "info" {
			infoSectionStart = bufferLength - reader.Buffered()
			infoSection, infoSectionExist = true, true
		}

		// get value
		nextBytes, _ = reader.Peek(1)
		nextByte = nextBytes[0]

		var value interface{}

		switch nextByte {
		case 'i':
			// IntType
			value = getInt(reader)
		case 'l':
			// ListType
			value = getList(reader)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			value = getBytes(reader)
		case 'd':
			// DictType
			value = getDict(reader)
		}
		dictData[key] = value
	}
}

func bencode_decode(data []byte) []interface{} {
	readerSize := 1024 * 1024 // 1MB Max Files Size
	reader := bufio.NewReaderSize(bytes.NewReader(data), readerSize)

	_, _ = reader.Peek(1)
	bufferLength = reader.Buffered()

	var decodedData []interface{}
	for {
		nextBytes, _ := reader.Peek(1)
		nextByte := nextBytes[0]

		switch nextByte {
		case 'i':
			// IntType
			decodedData = append(decodedData, getInt(reader))
		case 'l':
			// ListType
			decodedData = append(decodedData, getList(reader))
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			decodedData = append(decodedData, getBytes(reader))
		case 'd':
			decodedData = append(decodedData, getDict(reader))
		}

		if _, err := reader.Peek(1); err == io.EOF {
			break
		}
	}

	return decodedData

}

// Globals for Info Section
var infoSection, infoSectionExist bool
var infoSectionStart, infoSectionEnd, bufferLength int

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

func checkSingleMapInside(i []interface{}) bool {
	if len(i) == 1 {
		if _, ok := i[0].(map[string]interface{}); ok {
			return true
		}
	}
	return false
}

func main() {
	torrentFile := readTorrent("./small.torrent")
	log.Printf("%v character file", len(torrentFile))
	// log.Println(string(torrentFile))

	dataFromFile := bencode_decode(torrentFile)

	// Info Section Hash - Needs Work
	if infoSectionExist {
		log.Printf("info section start: %v, end: %v", infoSectionStart, infoSectionEnd)
		infoSectionBytes := torrentFile[infoSectionStart:infoSectionEnd]
		hash := sha1.Sum(infoSectionBytes)
		log.Printf("%x", hash)
	}

	// check data is from valid torrent file?
	// check data is one base level dict

	if !checkSingleMapInside(dataFromFile) {
		log.Fatalln("Data not in single dict")
	}

	torrentInterface := dataFromFile[0].(map[string]interface{})

	torrent := loadTorrentIntoStruct(torrentInterface)

	log.Println(torrent)

	numberOfPieces := (torrent.Info.Length / torrent.Info.PieceLength) + 1
	log.Println(numberOfPieces)

	var pieceHashes [][]byte
	pieceBuffer := torrent.Info.Pieces
	for range numberOfPieces {
		pieceHashes = append(pieceHashes, pieceBuffer[:20])
		pieceBuffer = pieceBuffer[20:]
	}
	log.Println(pieceBuffer)
	log.Println(pieceHashes)

	log.Println("Fin")

}
