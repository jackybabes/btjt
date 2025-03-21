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
	Bitfield       []byte
	BitfieldLength uint32
	Choked         bool
}

func NewPeerCompactIpv4(b [6]byte) Peer {
	// Create the Peer instance at the beginning
	p := Peer{}

	ipstr := fmt.Sprintf("%v.%v.%v.%v", b[0], b[1], b[2], b[3])
	p.IP = net.ParseIP(ipstr)
	p.Port = int(binary.BigEndian.Uint16(b[4:]))
	p.Name = fmt.Sprintf("%v:%v", ipstr, p.Port)

	peerID := make([]byte, 20)
	rand.Read(peerID)
	p.PeerID = peerID

	// Set peer address string
	p.peerAddress = fmt.Sprintf("%v:%v", p.IP.String(), p.Port)

	p.Choked = true

	return p
}

// func NewPeerCompactIpv6(b [16]byte) Peer {

func (p *Peer) Init(infoHash [20]byte) {

	// Create connection and store it in the struct
	conn := p.CreateConnection()
	if conn == nil {
		p.Alive = false
		return
	}
	p.Connection = conn // Store the connection in the struct instead of closing it

	// Create Handshake message
	handshake := createPeerHandshakeMessage(infoHash, p.PeerID)
	p.SendHandshake(handshake)

	// p.CreateConnection()

	// defer p.Connection.Close()
}

func createPeerHandshakeMessage(infoHash [20]byte, peerID []byte) []byte {
	handshakeLength := []byte{19}
	handshakeProtocolString := []byte("BitTorrent protocol")
	handshakeNullBytes := make([]byte, 8)
	handshakeInfoHash := infoHash[:]
	handshakePeerID := peerID
	handshake := slices.Concat(handshakeLength, handshakeProtocolString, handshakeNullBytes, handshakeInfoHash, handshakePeerID)
	return handshake
}

func (p *Peer) SendHandshake(handshakeMessage []byte) {
	_, err := p.Connection.Write(handshakeMessage)
	if err != nil {
		log.Println(err)
		p.Alive = false
		p.Connection.Close()
		return
	}

	log.Println("Handshake sent")

	response := make([]byte, 68)

	// p.Connection.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err = p.Connection.Read(response)
	if err != nil {
		log.Println(err)
		p.Alive = false
		p.Connection.Close()
		return
	}

	// // Remove the deadline after handshake
	// p.Connection.SetReadDeadline(time.Time{})

	log.Printf("Handshake response: %v", response)

	// Read bitfield message
	lengthBuf := make([]byte, 4)
	_, err = p.Connection.Read(lengthBuf)
	if err != nil {
		log.Println("Failed to read bitfield length:", err)
		p.Alive = false
		p.Connection.Close()
		return
	}

	p.BitfieldLength = binary.BigEndian.Uint32(lengthBuf)

	// First byte is message ID (5 for bitfield)
	messageBuf := make([]byte, p.BitfieldLength)
	_, err = p.Connection.Read(messageBuf)
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

func (p *Peer) SendInterested() error {
	p.SendMessage(Message{Type: INTERESTED})
	msg, err := p.ReceiveNextMessage()
	if err != nil {
		log.Println(err)
		p.Alive = false
		p.Connection.Close()
		return err
	}
	switch msg.Type {
	case UNCHOKE:
		log.Println("Peer unchoked")
		p.Choked = false
		return nil
	case CHOKE:
		p.Choked = true
		return fmt.Errorf("peer choked")
	default:
		return fmt.Errorf("expected interested message, got message ID: %d", msg.Type)
	}
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

func (p *Peer) Close() {
	if p.Connection != nil {
		p.Connection.Close()
		p.Connection = nil
	}
	p.Alive = false
}
