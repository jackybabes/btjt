package main

import (
	"log"
)

const (
	TORRENT_FILE_PATH = "./test_torrents/http.torrent"
	PORT              = 6881
	PEER_ID           = "jtbtjtbtjtbtjtbtjtbt"
	PROXY             = false
	DEFAULT_COMPACT   = 1
)

func main() {

	torrent := NewTorrent(TORRENT_FILE_PATH)

	torrent.createTrackers()

	for _, tracker := range torrent.Trackers {
		tracker.Announce(&torrent)
		if tracker.Alive {
			tracker.DecodeAnnounceResponse()
		}
	}

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
