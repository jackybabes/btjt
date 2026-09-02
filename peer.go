package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"slices"
	"strconv"
	"time"
)

// Number of block requests kept outstanding per peer.
const pipelineDepth = 5

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
	pending        int
}

func NewPeerCompactIpv4(b [6]byte) Peer {
	p := Peer{}

	ipstr := fmt.Sprintf("%v.%v.%v.%v", b[0], b[1], b[2], b[3])
	p.IP = net.ParseIP(ipstr)
	p.Port = int(binary.BigEndian.Uint16(b[4:]))
	p.Name = fmt.Sprintf("%v:%v", ipstr, p.Port)

	peerID := make([]byte, 20)
	rand.Read(peerID)
	p.PeerID = peerID

	p.peerAddress = fmt.Sprintf("%v:%v", p.IP.String(), p.Port)
	p.Choked = true

	return p
}

// NewPeerCompactIpv6 builds a Peer from an 18-byte BEP 7 compact entry:
// 16 bytes of IPv6 address followed by a 2-byte big-endian port.
func NewPeerCompactIpv6(b [18]byte) Peer {
	p := Peer{}

	p.IP = make(net.IP, net.IPv6len)
	copy(p.IP, b[0:16])
	p.Port = int(binary.BigEndian.Uint16(b[16:]))
	p.peerAddress = net.JoinHostPort(p.IP.String(), strconv.Itoa(p.Port))
	p.Name = p.peerAddress

	peerID := make([]byte, 20)
	rand.Read(peerID)
	p.PeerID = peerID

	p.Choked = true

	return p
}

// Init dials the peer and performs the BitTorrent handshake. On success
// p.Alive is set to true and p.Connection is ready for message exchange.
func (p *Peer) Init(infoHash [20]byte) {
	conn := p.CreateConnection()
	if conn == nil {
		p.Alive = false
		return
	}
	p.Connection = conn

	handshake := createPeerHandshakeMessage(infoHash, p.PeerID)
	if _, err := conn.Write(handshake); err != nil {
		log.Printf("%s: handshake write: %v", p.Name, err)
		p.Close()
		return
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	resp := make([]byte, 68)
	if _, err := io.ReadFull(conn, resp); err != nil {
		log.Printf("%s: handshake read: %v", p.Name, err)
		p.Close()
		return
	}
	conn.SetReadDeadline(time.Time{})

	if !bytes.Equal(resp[28:48], infoHash[:]) {
		log.Printf("%s: infohash mismatch", p.Name)
		p.Close()
		return
	}

	p.Alive = true
}

func createPeerHandshakeMessage(infoHash [20]byte, peerID []byte) []byte {
	handshakeLength := []byte{19}
	handshakeProtocolString := []byte("BitTorrent protocol")
	handshakeNullBytes := make([]byte, 8)
	handshakeInfoHash := infoHash[:]
	handshakePeerID := peerID
	return slices.Concat(handshakeLength, handshakeProtocolString, handshakeNullBytes, handshakeInfoHash, handshakePeerID)
}

// Download runs the peer's message loop until the torrent is complete or the
// connection drops.
func (p *Peer) Download(d *Download) {
	defer p.Close()

	if err := p.SendMessage(Message{Type: INTERESTED}); err != nil {
		return
	}

	for {
		if d.IsComplete() {
			return
		}

		p.Connection.SetReadDeadline(time.Now().Add(30 * time.Second))
		msg, err := p.ReceiveMessage()
		if err != nil {
			return
		}
		if msg == nil {
			continue // keep-alive
		}

		switch msg.Type {
		case BITFIELD:
			p.Bitfield = msg.Payload
		case HAVE:
			if len(msg.Payload) >= 4 {
				p.SetPiece(int(binary.BigEndian.Uint32(msg.Payload[:4])))
			}
		case UNCHOKE:
			p.Choked = false
		case CHOKE:
			p.Choked = true
		case PIECE:
			if len(msg.Payload) < 8 {
				break
			}
			pieceIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
			begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
			if p.pending > 0 {
				p.pending--
			}
			complete, err := d.SaveBlock(pieceIndex, begin, msg.Payload[8:])
			if err != nil {
				log.Printf("%s: %v", p.Name, err)
			}
			if complete {
				return
			}
		}

		// Keep the request pipeline topped up.
		for !p.Choked && p.pending < pipelineDepth {
			pieceIndex, block := d.NextRequest(p)
			if block == nil {
				break
			}
			if err := p.SendRequest(uint32(pieceIndex), uint32(block.Offset), uint32(block.Length)); err != nil {
				return
			}
			p.pending++
		}
	}
}

func (p *Peer) CreateConnection() net.Conn {
	conn, err := net.DialTimeout("tcp", p.peerAddress, 3*time.Second)
	if err != nil {
		return nil
	}
	return conn
}

func (p *Peer) Close() {
	if p.Connection != nil {
		p.Connection.Close()
		p.Connection = nil
	}
	p.Alive = false
}

// HasPiece reports whether the peer advertises piece pieceIndex in its bitfield.
func (p *Peer) HasPiece(pieceIndex int) bool {
	byteIndex := pieceIndex / 8
	bitIndex := 7 - (pieceIndex % 8) // bits count from the MSB

	if byteIndex < 0 || byteIndex >= len(p.Bitfield) {
		return false
	}
	return (p.Bitfield[byteIndex] & (1 << bitIndex)) != 0
}

// SetPiece marks a piece as available, growing the bitfield if necessary. Used
// to apply HAVE messages received after the initial bitfield.
func (p *Peer) SetPiece(pieceIndex int) {
	byteIndex := pieceIndex / 8
	bitIndex := 7 - (pieceIndex % 8)

	if byteIndex >= len(p.Bitfield) {
		grown := make([]byte, byteIndex+1)
		copy(grown, p.Bitfield)
		p.Bitfield = grown
	}
	p.Bitfield[byteIndex] |= 1 << bitIndex
}
