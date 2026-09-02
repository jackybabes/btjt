package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

const (
	PORT            = 6881
	PEER_ID         = "jtbtjtbtjtbtjtbtjtbt"
	PROXY           = false
	DEFAULT_COMPACT = 1
	BLOCK_SIZE      = 16 * 1024
)

func main() {
	outDir := flag.String("o", ".", "output directory")
	maxPeers := flag.Int("maxpeers", 40, "maximum number of peers to use")
	quiet := flag.Bool("quiet", false, "suppress verbose per-peer logging")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: btjt [-o dir] [-maxpeers n] [-quiet] <torrent-file>")
		os.Exit(2)
	}
	if *quiet {
		log.SetOutput(io.Discard)
	}

	torrent := NewTorrent(flag.Arg(0))
	torrent.initialiseTrackers()
	if len(torrent.Peers) == 0 {
		fmt.Fprintln(os.Stderr, "no peers returned by trackers")
		os.Exit(1)
	}

	storage, err := NewStorage(*outDir, &torrent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()
	torrent.Download.Storage = storage

	// Take up to maxPeers of the announced peers.
	peers := make([]*Peer, 0, len(torrent.Peers))
	for _, p := range torrent.Peers {
		peers = append(peers, p)
		if len(peers) >= *maxPeers {
			break
		}
	}

	// Handshake every peer concurrently.
	var hwg sync.WaitGroup
	for _, p := range peers {
		hwg.Add(1)
		go func(p *Peer) {
			defer hwg.Done()
			p.Init(torrent.InfoHash)
		}(p)
	}
	hwg.Wait()

	alive := peers[:0]
	for _, p := range peers {
		if p.Alive {
			alive = append(alive, p)
		}
	}
	fmt.Fprintf(os.Stderr, "connected to %d/%d peers\n", len(alive), len(peers))
	if len(alive) == 0 {
		os.Exit(1)
	}

	// Run a download loop per peer.
	start := time.Now()
	var dwg sync.WaitGroup
	for _, p := range alive {
		dwg.Add(1)
		go func(p *Peer) {
			defer dwg.Done()
			p.Download(torrent.Download)
		}(p)
	}

	// Progress reporter + stall watchdog.
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		last, stalls := 0, 0
		for range ticker.C {
			if torrent.Download.IsComplete() {
				return
			}
			done, total := torrent.Download.Progress()
			fmt.Fprintf(os.Stderr, "\rpieces %d/%d ", done, total)
			if done == last {
				if stalls++; stalls >= 30 { // ~60s without progress
					fmt.Fprintf(os.Stderr, "\nno progress for 60s, giving up (%d/%d pieces)\n", done, total)
					os.Exit(1)
				}
			} else {
				last, stalls = done, 0
			}
		}
	}()

	waitCh := make(chan struct{})
	go func() { dwg.Wait(); close(waitCh) }()

	select {
	case <-torrent.Download.Done:
	case <-waitCh:
	}

	done, total := torrent.Download.Progress()
	if torrent.Download.IsComplete() {
		fmt.Fprintf(os.Stderr, "\ndone: %d/%d pieces in %s\n", done, total, time.Since(start).Round(time.Second))
		return
	}
	fmt.Fprintf(os.Stderr, "\nincomplete: %d/%d pieces (all peers disconnected)\n", done, total)
	os.Exit(1)
}
