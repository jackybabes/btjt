package main

import (
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// Globals for Info Section
var infoSection, infoSectionExist bool
var infoSectionStart, infoSectionEnd, bufferLength int

func main() {
	PORT := 6881
	PEER_ID := "jtbtjtbtjtbtjtbtjtbt"

	torrentFile := readTorrent("./test_torrents/sample.torrent")
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
