package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
)

func parsePeerBytesIntoPeerList(peerBytes []byte) []Peer {
	var peerList []Peer
	peerNum := 0

	if len(peerList)%6 != 0 {
		log.Panicln("peer list not divisible by 6")
	}

	for range len(peerBytes) / 6 {
		offset := 6 * peerNum

		p := Peer{}

		p.IP = fmt.Sprintf("%v.%v.%v.%v", peerBytes[0+offset], peerBytes[1+offset], peerBytes[2+offset], peerBytes[3+offset])

		decimalPort := binary.BigEndian.Uint16(peerBytes[4+offset : 6+offset])

		p.Port = int(decimalPort)

		peerList = append(peerList, p)
		peerNum++

	}
	return peerList
}

func main() {
	PORT := 6881
	PEER_ID := "jtbtjtbtjtbtjtbtjtbt"
	PROXY := false

	torrentFile := readTorrent("./test_torrents/http.torrent")

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
	var torrentLengthTotal int
	if len(torrent.Info.Files) > 0 {
		// calc length
		for _, file := range torrent.Info.Files {
			torrentLengthTotal += file.Length
		}
	} else {
		torrentLengthTotal = torrent.Info.Length
	}
	numberOfPieces := (torrentLengthTotal / torrent.Info.PieceLength) + 1
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
	params.Set("left", strconv.Itoa(torrentLengthTotal))
	params.Set("downloaded", strconv.Itoa(0))
	params.Set("uploaded", strconv.Itoa(0))
	params.Set("compact", strconv.Itoa(1))

	finalURL := baseURL + "?" + params.Encode()
	log.Println(finalURL)

	resp, err := client.Get(finalURL)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()

	// Read and print response body
	log.Println(resp.Status)
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Response: %v...", string(body[:40]))

	trackerResponseDecoded, _ := bencode_decode([]byte(body))

	// assuming compact ipv4 stuff
	// https://www.bittorrent.org/beps/bep_0007.html

	trackerResponseUnpacked := UnpackTrackerResponse(trackerResponseDecoded)

	peerList := parsePeerBytesIntoPeerList(trackerResponseUnpacked.Peers)

	_ = peerList

	log.Println("Fin")

}
