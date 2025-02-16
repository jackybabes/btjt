package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
				// log.Println(reader.Buffered())
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
	readerSize := 1024 * 1024 * 4 // 4MB Max Files Size
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

func hashURLEncode(h [20]byte) string {
	// Note that all binary data in the URL (particularly info_hash and peer_id) must be properly escaped.
	// This means any byte not in the set 0-9, a-z, A-Z, '.', '-', '_' and '~', must be encoded using the "%nn" format,
	// where nn is the hexadecimal value of the byte. (See RFC1738 for details.)

	// For a 20-byte hash of \x12\x34\x56\x78\x9a\xbc\xde\xf1\x23\x45\x67\x89\xab\xcd\xef\x12\x34\x56\x78\x9a,
	// The right encoded form is %124Vx%9A%BC%DE%F1%23Eg%89%AB%CD%EF%124Vx%9A

	var encoded string
	for _, b := range h {

		if (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '.' || b == '-' || b == '_' || b == '~' {
			encoded += string(b)
			continue
		}

		hexOfByte := hex.EncodeToString([]byte{b})
		encoded += "%" + hexOfByte

	}
	return encoded
}

func main() {
	PORT := 6881
	PEER_ID := "jtbtjtbtjtbtjtbtjtbt"

	torrentFile := readTorrent("./http.torrent")
	log.Printf("%v character file", len(torrentFile))
	// log.Println(string(torrentFile))

	dataFromFile := bencode_decode(torrentFile)

	// Info Section Hash - Needs Work
	var infoHash [20]byte
	if infoSectionExist {
		log.Printf("info section start: %v, end: %v", infoSectionStart, infoSectionEnd)
		infoSectionBytes := torrentFile[infoSectionStart:infoSectionEnd]
		infoHash = sha1.Sum(infoSectionBytes)
		log.Printf("%x", infoHash)
	}

	// check data is from valid torrent file?
	// check data is one base level dict

	if !checkSingleMapInside(dataFromFile) {
		log.Fatalln("Data not in single dict")
	}

	torrentInterface := dataFromFile[0].(map[string]interface{})

	torrent := loadTorrentIntoStruct(torrentInterface)

	// log.Println(torrent)

	// Pieces Calc
	numberOfPieces := (torrent.Info.Length / torrent.Info.PieceLength) + 1
	log.Println(numberOfPieces)

	// Create slice of hashes
	var pieceHashes [][]byte
	pieceBuffer := torrent.Info.Pieces
	for range numberOfPieces {
		pieceHashes = append(pieceHashes, pieceBuffer[:20])
		pieceBuffer = pieceBuffer[20:]
	}
	// log.Println(pieceHashes)
	log.Printf("%x", pieceHashes[0])

	// Proxy through burp

	// Define the Burp Suite proxy URL
	proxyURL, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	// Create a custom Transport with the proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		// Skip TLS verification (needed for HTTPS if Burp is using its cert)
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Create an HTTP client with the custom transport
	client := &http.Client{
		Transport: transport,
	}

	// Inital Http Request
	baseURL := torrent.Announce
	params := url.Values{}
	params.Set("peer_id", PEER_ID)
	params.Set("info_hash", hashURLEncode(infoHash))
	params.Set("port", strconv.Itoa(PORT))
	params.Set("left", strconv.Itoa(torrent.Info.Length))
	params.Set("downloaded", strconv.Itoa(0))
	params.Set("uploaded", strconv.Itoa(0))
	params.Set("compact", strconv.Itoa(1))

	finalURL := baseURL + "?" + params.Encode()

	resp, err := client.Get(finalURL)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()

	// Read and print response body
	log.Println(resp.Status)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response:", string(body))

	log.Println("Fin")

}
