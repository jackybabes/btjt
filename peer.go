package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Peer struct {
	PeerID []byte
	IP     net.IP
	Port   int
	Alive  bool
}

func NewPeerCompactIpv4(b [6]byte) Peer {
	ipstr := fmt.Sprintf("%v.%v.%v.%v", b[0], b[1], b[2], b[3])
	ip := net.ParseIP(ipstr)
	port := int(binary.BigEndian.Uint16(b[4:]))

	p := Peer{IP: ip, Port: port}
	return p
}

// func parsePeerBytesIntoPeerList(peerBytes []byte) []Peer {
// 	var peerList []Peer
// 	peerNum := 0

// 	if len(peerList)%6 != 0 {
// 		log.Panicln("peer list not divisible by 6")
// 	}

// 	for range len(peerBytes) / 6 {
// 		offset := 6 * peerNum

// 		p := Peer{}

// 		p.IP = fmt.Sprintf("%v.%v.%v.%v", peerBytes[0+offset], peerBytes[1+offset], peerBytes[2+offset], peerBytes[3+offset])

// 		decimalPort := binary.BigEndian.Uint16(peerBytes[4+offset : 6+offset])

// 		p.Port = int(decimalPort)

// 		peerList = append(peerList, p)
// 		peerNum++

// 	}
// 	return peerList
// }

// func sendHandshakeToPeer(p Peer, infoHash [20]byte) []byte {
// 	// peerAddress := "188.212.112.163:32767"

// 	peerAddress := fmt.Sprintf("%v:%v", p.IP, p.Port)
// 	log.Println(peerAddress)

// 	conn, err := net.DialTimeout("tcp", peerAddress, time.Second)
// 	if err != nil {
// 		log.Println(err)
// 		return nil
// 	}
// 	defer conn.Close()

// 	log.Printf("connected to %v", peerAddress)

// 	peerID := make([]byte, 20)
// 	rand.Read(peerID)

// 	// length of the protocol string (BitTorrent protocol) which is 19 (1 byte)
// 	handshakeLength := []byte{19}
// 	// the string BitTorrent protocol (19 bytes)
// 	handshakeProtocolString := []byte("BitTorrent protocol")
// 	// eight reserved bytes, which are all set to zero (8 bytes)
// 	handshakeNullBytes := make([]byte, 8)
// 	// sha1 infohash (20 bytes) (NOT the hexadecimal representation, which is 40 bytes long)
// 	handshakeInfoHash := infoHash[:]
// 	// peer id (20 bytes) (generate 20 random byte values)
// 	handshakePeerID := peerID

// 	handshake := slices.Concat(handshakeLength, handshakeProtocolString, handshakeNullBytes, handshakeInfoHash, handshakePeerID)

// 	log.Printf("Handshake len : %v", len(handshake))

// 	_, err = conn.Write(handshake)
// 	if err != nil {
// 		log.Println(err)
// 		return nil
// 	}

// 	log.Println("Handshake sent")

// 	response := make([]byte, 68)

// 	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
// 	n, err := conn.Read(response)
// 	if err != nil {
// 		log.Println(err)
// 		return nil
// 	}

// 	log.Println(n)

// 	log.Printf("%x", response[68-20:])
// 	return response
// }
