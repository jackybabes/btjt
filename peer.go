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

	// Create connection and store it in the struct
	conn := p.CreateConnection()
	if conn == nil {
		p.Alive = false
		return
	}
	p.Connection = conn // Store the connection in the struct instead of closing it

	// Create Handshake message
	p.createHandshakeMessage(infoHash)
	p.SendHandshake()

	// p.CreateConnection()

	// defer p.Connection.Close()
}

func (p *Peer) sendInterested() {
	messageLen := []byte{0, 0, 0, 1}
	messageID := []byte{2}
	message := slices.Concat(messageLen, messageID)
	log.Printf("Sending interested: %v", message)
	_, err := p.Connection.Write(message)
	if err != nil {
		log.Println(err)
	}
}

func (p *Peer) receiveUnchoke() {
	response := make([]byte, 5)
	_, err := p.Connection.Read(response)
	if err != nil {
		log.Println(err)
		p.Alive = false
		p.Connection.Close()
		return
	}
	log.Printf("Received unchoke: %v", response)
}

func (p *Peer) sendRequest(peiceIndexInt int, blockOffsetInt int) {
	// message length
	messageLen := []byte{0, 0, 0, 13}

	// message ID
	messageID := []byte{6}

	// piece index
	pieceIndex := make([]byte, 4)
	binary.BigEndian.PutUint32(pieceIndex, uint32(peiceIndexInt))

	// block offset
	blockOffset := make([]byte, 4)
	binary.BigEndian.PutUint32(blockOffset, uint32(blockOffsetInt))

	// block length
	blockLength := make([]byte, 4)
	binary.BigEndian.PutUint32(blockLength, uint32(BLOCK_SIZE))

	// message
	message := slices.Concat(messageLen, messageID, pieceIndex, blockOffset, blockLength)

	log.Printf("Sending request: %v", message)
	_, err := p.Connection.Write(message)
	if err != nil {
		log.Println(err)
	}
}

func (p *Peer) receivePiece() {
	log.Printf("Receiving piece")

	// First read the message length (4 bytes)
	lengthBuf := make([]byte, 4)
	_, err := p.Connection.Read(lengthBuf)
	if err != nil {
		log.Println("Failed to read piece message length:", err)
		p.Alive = false
		p.Connection.Close()
		return
	}

	messageLength := binary.BigEndian.Uint32(lengthBuf)
	log.Printf("Message length: %d", messageLength)

	// Read the actual message
	response := make([]byte, messageLength)
	_, err = p.Connection.Read(response)
	if err != nil {
		log.Println("Failed to read piece message:", err)
		p.Alive = false
		p.Connection.Close()
		return
	}

	if response[0] != 7 {
		log.Printf("Expected piece message (ID=7), got message ID: %d", response[0])
		return
	}

	// Parse piece message
	pieceIndex := binary.BigEndian.Uint32(response[1:5])
	blockOffset := binary.BigEndian.Uint32(response[5:9])
	blockData := response[9:]

	log.Printf("Piece index: %d, Block offset: %d, Block data length: %d",
		pieceIndex, blockOffset, len(blockData))
	log.Printf("First few bytes of block data: %v", blockData[:min(10, len(blockData))])
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
	_, err := p.Connection.Write(p.handshakeMsg)
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
