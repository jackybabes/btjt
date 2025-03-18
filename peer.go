package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"slices"
	"time"
)

type Peer struct {
	PeerID         []byte
	Name           string
	IP             net.IP
	Port           int
	Alive          bool
	Connection     net.Conn
	peerAddress    string
	handshakeMsg   []byte
	Bitfield       []byte
	BitfieldLength uint32
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

// func NewPeerCompactIpv6(b [16]byte) Peer {

func (p *Peer) Init(infoHash [20]byte) {
	// Set peer address string
	p.peerAddress = fmt.Sprintf("%v:%v", p.IP.String(), p.Port)

	// Create Handshake message
	p.createHandshakeMessage(infoHash)
	p.SendHandshake()

	// p.CreateConnection()

	// defer p.Connection.Close()
}

func (p *Peer) createHandshakeMessage(infoHash [20]byte) {
	handshakeLength := []byte{19}
	handshakeProtocolString := []byte("BitTorrent protocol")
	handshakeNullBytes := make([]byte, 8)
	handshakeInfoHash := infoHash[:]
	handshakePeerID := p.PeerID

	handshake := slices.Concat(handshakeLength, handshakeProtocolString, handshakeNullBytes, handshakeInfoHash, handshakePeerID)

	p.handshakeMsg = handshake
}

func (p *Peer) SendHandshake() {
	// Create connection and store it in the struct
	conn := p.CreateConnection()
	if conn == nil {
		p.Alive = false
		return
	}
	p.Connection = conn // Store the connection in the struct instead of closing it

	_, err := conn.Write(p.handshakeMsg)
	if err != nil {
		log.Println(err)
		p.Alive = false
		p.Connection.Close()
		return
	}

	log.Println("Handshake sent")

	response := make([]byte, 68)

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Read(response)
	if err != nil {
		log.Println(err)
		p.Alive = false
		p.Connection.Close()
		return
	}

	// Remove the deadline after handshake
	conn.SetReadDeadline(time.Time{})

	log.Printf("Handshake response: %v", response)

	// Read bitfield message
	lengthBuf := make([]byte, 4)
	_, err = conn.Read(lengthBuf)
	if err != nil {
		log.Println("Failed to read bitfield length:", err)
		p.Alive = false
		p.Connection.Close()
		return
	}

	p.BitfieldLength = binary.BigEndian.Uint32(lengthBuf)

	// First byte is message ID (5 for bitfield)
	messageBuf := make([]byte, p.BitfieldLength)
	_, err = conn.Read(messageBuf)
	if err != nil {
		log.Println("Failed to read bitfield:", err)
		p.Alive = false
		p.Connection.Close()
		return
	}

	if messageBuf[0] != 5 { // 5 is the ID for bitfield messages
		log.Println("Expected bitfield message, got message ID:", messageBuf[0])
		p.Alive = false
		p.Connection.Close()
		return
	}

	p.Bitfield = messageBuf[1:] // The actual bitfield data starts after the message ID
	log.Printf("Received bitfield of length %d: %v", len(p.Bitfield), p.Bitfield)
	p.Alive = true // Mark as alive if we got here successfully
}

func (p *Peer) CreateConnection() net.Conn {
	conn, err := net.DialTimeout("tcp", p.peerAddress, time.Second)
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Printf("connected to %v", p.peerAddress)
	return conn
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

func (p *Peer) Close() {
	if p.Connection != nil {
		p.Connection.Close()
		p.Connection = nil
	}
	p.Alive = false
}
