package main

import (
	"log"
)

func main() {
	// PORT := 6881
	// PEER_ID := "jtbtjtbtjtbtjtbtjtbt"
	// PROXY := false
	TORRENT_FILE_PATH := "./test_torrents/http.torrent"

	torrent := NewTorrent(TORRENT_FILE_PATH)

	log.Println(torrent)

	// // Pieces Calc -  broken
	// // When multiple files, length is store in file dict
	// var torrentLengthTotal int
	// if len(torrent.Info.Files) > 0 {
	// 	// calc length
	// 	for _, file := range torrent.Info.Files {
	// 		torrentLengthTotal += file.Length
	// 	}
	// } else {
	// 	torrentLengthTotal = torrent.Info.Length
	// }
	// numberOfPieces := (torrentLengthTotal / torrent.Info.PieceLength) + 1
	// log.Printf("number of pieces: %v", numberOfPieces)

	// // Create slice of hashes
	// var pieceHashes [][]byte
	// pieceBuffer := torrent.Info.Pieces
	// for range numberOfPieces {
	// 	pieceHashes = append(pieceHashes, pieceBuffer[:20])
	// 	pieceBuffer = pieceBuffer[20:]
	// }
	// // log.Println(pieceHashes)
	// log.Printf("%x", pieceHashes[0])

	// client := getClient(PROXY)

	// // Inital Http Request
	// baseURL := torrent.Announce
	// params := url.Values{}
	// params.Set("peer_id", PEER_ID)
	// // Does not work here as double url encodes %
	// params.Set("info_hash", string(infoHash[:]))
	// params.Set("port", strconv.Itoa(PORT))
	// params.Set("left", strconv.Itoa(torrentLengthTotal))
	// params.Set("downloaded", strconv.Itoa(0))
	// params.Set("uploaded", strconv.Itoa(0))
	// params.Set("compact", strconv.Itoa(1))

	// finalURL := baseURL + "?" + params.Encode()
	// log.Println(finalURL)

	// resp, err := client.Get(finalURL)
	// if err != nil {
	// 	fmt.Println("Request error:", err)
	// 	return
	// }
	// defer resp.Body.Close()

	// // Read and print response body
	// log.Println(resp.Status)
	// body, _ := io.ReadAll(resp.Body)
	// log.Printf("Response: %v...", string(body[:40]))

	// trackerResponseDecoded, _ := bencode_decode([]byte(body))

	// // assuming compact ipv4 stuff
	// // https://www.bittorrent.org/beps/bep_0007.html

	// trackerResponseUnpacked := UnpackTrackerResponse(trackerResponseDecoded)

	// peerList := parsePeerBytesIntoPeerList(trackerResponseUnpacked.Peers)

	// _ = peerList

	// // handshake

	// handshakePeer := peerList[0]

	// sendHandshakeToPeer(handshakePeer, infoHash)

	log.Println("Fin")

}
