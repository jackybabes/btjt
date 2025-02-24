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
	torrent.initialiseTrackers()

	log.Println("Fin")

}

// // handshake

// handshakePeer := peerList[0]

// sendHandshakeToPeer(handshakePeer, infoHash)
