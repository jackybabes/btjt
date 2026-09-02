package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"net/url"
	"time"
)

// BEP 15 - UDP tracker protocol.

const udpProtocolID = 0x41727101980

const (
	udpActionConnect  = 0
	udpActionAnnounce = 1
	udpActionError    = 3
)

const (
	udpTimeout  = 5 * time.Second
	udpAttempts = 2
)

// announceUDP performs a BEP 15 connect + announce and writes the result
// straight into tracker.CompactData / tracker.Interval, so the rest of the
// (HTTP-oriented) pipeline can treat it exactly like a compact HTTP response.
func (tracker *Tracker) announceUDP(t *Torrent) {
	tracker.Alive = false

	u, err := url.Parse(tracker.URL)
	if err != nil {
		log.Printf("udp tracker: bad url %q: %v", tracker.URL, err)
		return
	}

	conn, err := net.DialTimeout("udp", u.Host, udpTimeout)
	if err != nil {
		log.Printf("udp tracker %s: dial: %v", u.Host, err)
		return
	}
	defer conn.Close()

	connID, err := udpConnect(conn)
	if err != nil {
		log.Printf("udp tracker %s: connect: %v", u.Host, err)
		return
	}

	interval, seeders, leechers, peers, err := udpAnnounce(conn, connID, t)
	if err != nil {
		log.Printf("udp tracker %s: announce: %v", u.Host, err)
		return
	}

	tracker.Compact = 1
	tracker.CompactData = TrackerResponseCompact{
		Interval:   interval,
		Complete:   seeders,
		Incomplete: leechers,
		Peers:      peers,
	}
	tracker.Interval = interval
	tracker.Alive = true
	log.Printf("udp tracker %s: %d seeders, %d leechers, %d peers", u.Host, seeders, leechers, len(peers)/6)
}

// udpRoundTrip sends req and returns the first datagram received, retrying a
// few times on timeout. UDP is a message protocol so one Read is one packet.
func udpRoundTrip(conn net.Conn, req []byte) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < udpAttempts; attempt++ {
		if _, err := conn.Write(req); err != nil {
			lastErr = err
			continue
		}
		conn.SetReadDeadline(time.Now().Add(udpTimeout))
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			lastErr = err
			continue
		}
		return buf[:n], nil
	}
	return nil, lastErr
}

func udpConnect(conn net.Conn) (uint64, error) {
	txID := rand.Uint32()

	req := make([]byte, 16)
	binary.BigEndian.PutUint64(req[0:8], udpProtocolID)
	binary.BigEndian.PutUint32(req[8:12], udpActionConnect)
	binary.BigEndian.PutUint32(req[12:16], txID)

	resp, err := udpRoundTrip(conn, req)
	if err != nil {
		return 0, err
	}
	if len(resp) < 16 {
		return 0, fmt.Errorf("connect response too short: %d bytes", len(resp))
	}

	action := binary.BigEndian.Uint32(resp[0:4])
	if binary.BigEndian.Uint32(resp[4:8]) != txID {
		return 0, fmt.Errorf("connect: transaction id mismatch")
	}
	if action == udpActionError {
		return 0, fmt.Errorf("tracker error: %s", string(resp[8:]))
	}
	if action != udpActionConnect {
		return 0, fmt.Errorf("connect: unexpected action %d", action)
	}
	return binary.BigEndian.Uint64(resp[8:16]), nil
}

func udpAnnounce(conn net.Conn, connID uint64, t *Torrent) (interval, seeders, leechers int, peers []byte, err error) {
	txID := rand.Uint32()

	req := make([]byte, 98)
	binary.BigEndian.PutUint64(req[0:8], connID)
	binary.BigEndian.PutUint32(req[8:12], udpActionAnnounce)
	binary.BigEndian.PutUint32(req[12:16], txID)
	copy(req[16:36], t.InfoHash[:])
	copy(req[36:56], []byte(PEER_ID))
	binary.BigEndian.PutUint64(req[56:64], uint64(t.Downloaded))
	binary.BigEndian.PutUint64(req[64:72], uint64(t.Left))
	binary.BigEndian.PutUint64(req[72:80], uint64(t.Uploaded))
	binary.BigEndian.PutUint32(req[80:84], 2)             // event: started
	binary.BigEndian.PutUint32(req[84:88], 0)             // IP: let the tracker use our source address
	binary.BigEndian.PutUint32(req[88:92], rand.Uint32()) // key
	binary.BigEndian.PutUint32(req[92:96], ^uint32(0))    // num_want: -1 (tracker's choice)
	binary.BigEndian.PutUint16(req[96:98], uint16(PORT))

	resp, err := udpRoundTrip(conn, req)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	if len(resp) < 20 {
		return 0, 0, 0, nil, fmt.Errorf("announce response too short: %d bytes", len(resp))
	}

	action := binary.BigEndian.Uint32(resp[0:4])
	if binary.BigEndian.Uint32(resp[4:8]) != txID {
		return 0, 0, 0, nil, fmt.Errorf("announce: transaction id mismatch")
	}
	if action == udpActionError {
		return 0, 0, 0, nil, fmt.Errorf("tracker error: %s", string(resp[8:]))
	}
	if action != udpActionAnnounce {
		return 0, 0, 0, nil, fmt.Errorf("announce: unexpected action %d", action)
	}

	interval = int(binary.BigEndian.Uint32(resp[8:12]))
	leechers = int(binary.BigEndian.Uint32(resp[12:16]))
	seeders = int(binary.BigEndian.Uint32(resp[16:20]))

	// Remaining bytes are 6-byte IPv4+port entries, same layout as the compact
	// HTTP response. Trim any trailing partial entry defensively.
	body := resp[20:]
	body = body[:len(body)/6*6]
	peers = append([]byte(nil), body...)

	return interval, seeders, leechers, peers, nil
}
