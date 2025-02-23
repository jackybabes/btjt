package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type Tracker struct {
	URL            string
	Tier           int
	Compact        int
	Alive          bool
	BencodedData   []byte
	debugInterface map[string]interface{}
}

func NewTracker(url string, tier int) *Tracker {
	return &Tracker{URL: url, Tier: tier, Compact: DEFAULT_COMPACT}
}

func (tracker *Tracker) Announce(t *Torrent) {

	log.Printf("URL: %v", tracker.URL)

	baseURL := tracker.URL
	params := url.Values{}
	params.Set("peer_id", PEER_ID)
	params.Set("info_hash", string(t.InfoHash[:]))
	params.Set("port", strconv.Itoa(PORT))
	params.Set("left", strconv.Itoa(t.Left))
	params.Set("downloaded", strconv.Itoa(t.Downloaded))
	params.Set("uploaded", strconv.Itoa(t.Uploaded))
	params.Set("compact", strconv.Itoa(tracker.Compact))

	finalURL := baseURL + "?" + params.Encode()
	// log.Println(finalURL)

	resp, err := http.Get(finalURL)
	if err != nil {
		log.Println("Request error:", err)
		tracker.Alive = false
		return
	}
	defer resp.Body.Close()

	// Read and print response body
	// log.Println(resp.Status)
	if resp.Status != "200 OK" {
		log.Println("Status Error:", resp.Status)
		tracker.Alive = false
		return
	}

	tracker.Alive = true
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Response: %v...", string(body[:40]))
	tracker.BencodedData = body
}

func (tracker *Tracker) DecodeAnnounceResponse() {
	tracker.debugInterface = bencode_decode(tracker.BencodedData)

	if tracker.Compact = checkCompact(tracker.debugInterface); tracker.Compact < 0 {
		log.Println("Could not determine compactness, marking tracker Dead")
		tracker.Alive = false
		return
	}

	log.Println(len(tracker.debugInterface))
}

func checkCompact(m map[string]interface{}) int {
	switch m["peers"].(type) {
	case []byte:
		return 1
	case map[string]interface{}:
		return 0
	default:
		// log.Println(v)
		return -1

	}
}
