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

func getClient(proxy bool) *http.Client {
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

	var client *http.Client

	if proxy {
		// Create an HTTP client with the custom transport
		client = &http.Client{
			Transport: transport,
		}
	} else {
		client = &http.Client{}
	}

	return client

}

func main() {
	PORT := 6881
	PEER_ID := "jtbtjtbtjtbtjtbtjtbt"
	PROXY := false

	torrentFile := readTorrent("./test_torrents/ubuntu.torrent")

	// Check file is dict
	if torrentFile[0] != 'd' {
		log.Panicln("Torrent File not dict")
	}

	// bencode decode torrent file
	log.Printf("%v character torrent file", len(torrentFile))
	dataFromFile, infoSectionBytes := bencode_decode(torrentFile)

	// Info Section Hash
	infoHash := sha1.Sum(infoSectionBytes)
	log.Printf("%x", infoHash)

	torrent := loadTorrentIntoStruct(dataFromFile)

	// Pieces Calc -  broken
	// When multiple files, length is store in file dict
	numberOfPieces := (torrent.Info.Length / torrent.Info.PieceLength) + 1
	log.Printf("number of pieces: %v", numberOfPieces)

	// Create slice of hashes
	var pieceHashes [][]byte
	pieceBuffer := torrent.Info.Pieces
	for range numberOfPieces {
		pieceHashes = append(pieceHashes, pieceBuffer[:20])
		pieceBuffer = pieceBuffer[20:]
	}
	// log.Println(pieceHashes)
	log.Printf("%x", pieceHashes[0])

	client := getClient(PROXY)

	// Inital Http Request
	baseURL := torrent.Announce
	params := url.Values{}
	params.Set("peer_id", PEER_ID)
	// Does not work here as double url encodes %
	params.Set("info_hash", string(infoHash[:]))
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
