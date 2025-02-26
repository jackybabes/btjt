package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type Peer struct {
	PeerID     []byte
	Name       string
	IP         net.IP
	Port       int
	Alive      bool
	Connection net.Conn
}

func NewPeerCompactIpv4(b [6]byte) Peer {
	ipstr := fmt.Sprintf("%v.%v.%v.%v", b[0], b[1], b[2], b[3])
	ip := net.ParseIP(ipstr)
	port := int(binary.BigEndian.Uint16(b[4:]))

	peerstr := fmt.Sprintf("%v:%v", ipstr, port)

	peerID := make([]byte, 20)
	rand.Read(peerID)

	p := Peer{IP: ip, Port: port, Name: peerstr, PeerID: peerID}
	return p
}

func (p *Peer) Init() {
	p.CreateConnection()

	defer p.Connection.Close()
}

func (p *Peer) CreateConnection() {

	peerAddress := fmt.Sprintf("%v:%v", p.IP.String(), p.Port)
	conn, err := net.DialTimeout("tcp", peerAddress, time.Second)
	if err != nil {
		log.Println(err)
		return
	}
	p.Connection = conn
	log.Printf("connected to %v", peerAddress)

}

// func (p *Peer) SendHandshake() {
// 	// length of the protocol string (BitTorrent protocol) which is 19 (1 byte)
// 	handshakeLength := []byte{19}
// 	// the string BitTorrent protocol (19 bytes)
// 	handshakeProtocolString := []byte("BitTorrent protocol")
// 	// eight reserved bytes, which are all set to zero (8 bytes)
// 	handshakeNullBytes := make([]byte, 8)
// 	// sha1 infohash (20 bytes) (NOT the hexadecimal representation, which is 40 bytes long)
// 	handshakeInfoHash := infoHash[:]
// 	// peer id (20 bytes) (generate 20 random byte values)
// 	handshakePeerID := p.PeerID

// 	handshake := slices.Concat(handshakeLength, handshakeProtocolString, handshakeNullBytes, handshakeInfoHash, handshakePeerID)

// 	log.Printf("Handshake len : %v", len(handshake))

// 	_, err := p.Connection.Write(handshake)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	log.Println("Handshake sent")

// 	response := make([]byte, 68)

// 	p.Connection.SetReadDeadline(time.Now().Add(3 * time.Second))
// 	n, err := p.Connection.Read(response)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	log.Println(n)

// 	log.Printf("%x", response[68-20:])
// 	return
// }
