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
	CompactData    TrackerResponseCompact
}

// type TrackerResponseExpanded struct {
// 	FailureReason  string
// 	WarningMessage string
// 	Interval       int
// 	MinInterval    int
// 	TrackerID      []byte
// 	Complete       int
// 	Incomplete     int
// 	PeersDict      []Peer
// 	PeersBinary    []byte
// }

type TrackerResponseCompact struct {
	Interval   int
	Complete   int
	Incomplete int
	Peers      []byte
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

	announceResponseCompactness := checkCompact(tracker.debugInterface["peers"])
	if announceResponseCompactness < 0 {
		log.Println("Could not determine compactness, marking tracker Dead")
		tracker.Alive = false
		return
	}
	tracker.Compact = announceResponseCompactness

	// decode compact repsonse
	if tracker.Compact == 1 {
		populateStruct(&tracker.CompactData, tracker.debugInterface)
	}
	if tracker.Compact == 0 {
		log.Fatalln("Tracker Reponse not compact. Not supported yet.")
	}
}

func checkCompact(m interface{}) int {
	switch m.(type) {
	case []byte:
		return 1
	case map[string]interface{}:
		return 0
	default:
		// log.Println(v)
		return -1

	}
}
