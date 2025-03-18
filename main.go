package main

import (
	"log"
	"sync"
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

	// for testing reduce the number of peers
	if len(torrent.Peers) > 10 {
		// Create a temporary map to store the first 10 peers
		tempPeers := make(map[string]*Peer)
		count := 0
		for k, v := range torrent.Peers {
			if count >= 10 {
				break
			}
			tempPeers[k] = v
			count++
		}
		torrent.Peers = tempPeers
	}

	// init peers using goroutines
	var wg sync.WaitGroup
	for _, peer := range torrent.Peers {
		wg.Add(1)
		go func(p *Peer) {
			defer wg.Done()
			log.Printf("peer: %v", p.Name)
			p.Init(torrent.InfoHash)
		}(peer)
	}
	wg.Wait()

	// Remove peers that are not alive for testing
	for key, peer := range torrent.Peers {
		if !peer.Alive {
			delete(torrent.Peers, key)
		}
	}
	log.Printf("Remaining alive peers: %d", len(torrent.Peers))

	log.Println("Fin")

}

// // handshake

// handshakePeer := peerList[0]

// sendHandshakeToPeer(handshakePeer, infoHash)
