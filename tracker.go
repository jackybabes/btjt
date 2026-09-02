package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Tracker struct {
	URL            string
	Tier           int
	Compact        int
	Alive          bool
	BencodedData   []byte
	debugInterface map[string]any
	CompactData    TrackerResponseCompact
	Interval       int
	Peers          []Peer
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

func NewTracker(url string, tier int) Tracker {
	return Tracker{URL: url, Tier: tier, Compact: DEFAULT_COMPACT}
}

func (tracker *Tracker) InitialiseTracker(t *Torrent) {
	tracker.Announce(t)
	if tracker.Alive {
		tracker.DecodeAnnounceResponse()
	}
	if tracker.Alive {
		tracker.GetInterval()
	}
	if tracker.Alive {
		tracker.CreatePeerList()
	}
	// add peers to torrent Peers map

	for i := range tracker.Peers {
		peer := &tracker.Peers[i]
		if _, exists := t.Peers[peer.Name]; !exists {
			t.Peers[peer.Name] = peer
		}
	}
}

func (tracker *Tracker) isUDP() bool {
	return strings.HasPrefix(tracker.URL, "udp://")
}

func (tracker *Tracker) Announce(t *Torrent) {
	log.Printf("URL: %v", tracker.URL)

	if tracker.isUDP() {
		tracker.announceUDP(t)
		return
	}

	baseURL := tracker.URL
	params := url.Values{}
	params.Set("peer_id", PEER_ID)
	params.Set("info_hash", string(t.InfoHash[:]))
	params.Set("port", strconv.Itoa(PORT))
	params.Set("left", strconv.Itoa(t.Left))
	params.Set("downloaded", strconv.Itoa(t.Downloaded))
	params.Set("uploaded", strconv.Itoa(t.Uploaded))
	params.Set("compact", strconv.Itoa(tracker.Compact))
	params.Set("numwant", "100")
	params.Set("event", "started")

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
	if resp.StatusCode != http.StatusOK {
		log.Println("Status Error:", resp.Status)
		tracker.Alive = false
		return
	}

	tracker.Alive = true
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Response: %v...", string(body[:min(40, len(body))]))
	tracker.BencodedData = body
}

func (tracker *Tracker) DecodeAnnounceResponse() {
	if tracker.isUDP() {
		return // announceUDP already populated CompactData
	}

	tracker.debugInterface = bencode_decode(tracker.BencodedData)

	if reason, ok := tracker.debugInterface["failure reason"]; ok {
		log.Printf("tracker %s refused: %s", tracker.URL, bencodeString(reason))
		tracker.Alive = false
		return
	}

	announceResponseCompactness := checkCompact(tracker.debugInterface["peers"])
	if announceResponseCompactness < 0 {
		log.Println("Could not determine compactness, marking tracker Dead")
		tracker.Alive = false
		return
	}
	tracker.Compact = announceResponseCompactness

	log.Printf("tracker %s: complete=%v incomplete=%v peers=%dB peers6=%dB",
		tracker.URL, tracker.debugInterface["complete"], tracker.debugInterface["incomplete"],
		byteLen(tracker.debugInterface["peers"]), byteLen(tracker.debugInterface["peers6"]))

	// decode compact repsonse
	if tracker.Compact == 1 {
		populateStruct(&tracker.CompactData, tracker.debugInterface)
	}
	if tracker.Compact == 0 {
		tracker.Alive = false
		log.Println("Tracker Reponse not compact. Not supported yet.")
	}
}

// bencodeString renders a decoded bencode string value (this decoder yields
// []byte) for logging.
func bencodeString(v any) string {
	switch s := v.(type) {
	case []byte:
		return string(s)
	case string:
		return s
	default:
		return ""
	}
}

// byteLen reports the length of a decoded bencode string value, or 0.
func byteLen(v any) int {
	if b, ok := v.([]byte); ok {
		return len(b)
	}
	return 0
}

func checkCompact(m any) int {
	switch m.(type) {
	case []byte:
		return 1
	case map[string]any:
		return 0
	default:
		return -1
	}
}

func (tracker *Tracker) GetInterval() {
	// Parse Compact
	if tracker.Compact == 1 {
		tracker.Interval = tracker.CompactData.Interval
	}
	// Parse Expanded
	// if tracker.Compact == 0 {

	// }

}

func (tracker *Tracker) CreatePeerList() {
	// Compact IPv4 peers: "peers" as a string of 6-byte entries.
	if tracker.Compact == 1 && len(tracker.CompactData.Peers)%6 == 0 {
		for _, b := range ConvertTo6ByteSlices(tracker.CompactData.Peers) {
			tracker.Peers = append(tracker.Peers, NewPeerCompactIpv4(b))
		}
	}

	// Compact IPv6 peers (BEP 7): "peers6" as a string of 18-byte entries.
	// Dual-stack trackers return most of their peers here when we announce
	// over IPv6.
	if raw, ok := tracker.debugInterface["peers6"].([]byte); ok {
		for _, b := range ConvertTo18ByteSlices(raw) {
			tracker.Peers = append(tracker.Peers, NewPeerCompactIpv6(b))
		}
	}

	log.Printf("tracker %s: parsed %d peers", tracker.URL, len(tracker.Peers))
}
