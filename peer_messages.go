package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"slices"
)

// maxMessageSize guards against a desynced stream claiming an absurd length.
const maxMessageSize = 1 << 20

// peer messages
// All non-keepalive messages start with a single byte which gives their type.

// The possible values are:

// 0 - choke
// 1 - unchoke
// 2 - interested
// 3 - not interested
// 4 - have
// 5 - bitfield
// 6 - request
// 7 - piece
// 8 - cancel
// 'choke', 'unchoke', 'interested', and 'not interested' have no payload.

// Message types
const (
	CHOKE          = 0
	UNCHOKE        = 1
	INTERESTED     = 2
	NOT_INTERESTED = 3
	HAVE           = 4
	BITFIELD       = 5
	REQUEST        = 6
	PIECE          = 7
	CANCEL         = 8
)

// messageTypeToString converts a message type byte to its string representation
func messageTypeToString(msgType byte) string {
	switch msgType {
	case CHOKE:
		return "CHOKE"
	case UNCHOKE:
		return "UNCHOKE"
	case INTERESTED:
		return "INTERESTED"
	case NOT_INTERESTED:
		return "NOT_INTERESTED"
	case HAVE:
		return "HAVE"
	case BITFIELD:
		return "BITFIELD"
	case REQUEST:
		return "REQUEST"
	case PIECE:
		return "PIECE"
	case CANCEL:
		return "CANCEL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", msgType)
	}
}

// Message represents a BitTorrent peer message
type Message struct {
	Type    byte
	Payload []byte
}

// SendMessage sends a message to the peer
func (p *Peer) SendMessage(msg Message) error {
	// Calculate message length (1 byte for type + payload length)
	length := 1 + len(msg.Payload)

	// Create length prefix (4 bytes)
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(length))

	// Combine length, type, and payload
	message := slices.Concat(lengthBytes, []byte{msg.Type}, msg.Payload)

	_, err := p.Connection.Write(message)
	if err != nil {
		log.Printf("Failed to send message type %s: %v", messageTypeToString(msg.Type), err)
		p.Alive = false
		p.Connection.Close()
		return err
	}
	return nil
}

// ReceiveNextMessage reads messages until it gets a non-keep-alive message
func (p *Peer) ReceiveNextMessage() (*Message, error) {
	for {
		msg, err := p.ReceiveMessage()
		if err != nil {
			return nil, err
		}

		// If msg is nil, it was a keep-alive message, continue reading
		if msg == nil {
			log.Printf("Received keep-alive message, continuing...")
			continue
		}

		// Return any non-keep-alive message
		return msg, nil
	}
}

// ReceiveMessage reads one length-prefixed message from the peer. It returns
// (nil, nil) for a keep-alive.
func (p *Peer) ReceiveMessage() (*Message, error) {
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(p.Connection, lengthBuf); err != nil {
		p.Alive = false
		p.Connection.Close()
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, nil // keep-alive
	}
	if length > maxMessageSize {
		p.Alive = false
		p.Connection.Close()
		return nil, fmt.Errorf("peer sent oversized message: %d bytes", length)
	}

	messageBuf := make([]byte, length)
	if _, err := io.ReadFull(p.Connection, messageBuf); err != nil {
		p.Alive = false
		p.Connection.Close()
		return nil, err
	}

	return &Message{
		Type:    messageBuf[0],
		Payload: messageBuf[1:],
	}, nil
}

func (p *Peer) SendRequest(pieceIndex, blockOffset, blockLength uint32) error {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], pieceIndex)
	binary.BigEndian.PutUint32(payload[4:8], blockOffset)
	binary.BigEndian.PutUint32(payload[8:12], blockLength)

	return p.SendMessage(Message{
		Type:    REQUEST,
		Payload: payload,
	})
}
