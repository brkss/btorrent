package handshake

import (
	"fmt"
	"io"
)

// Handshake is a spcial message that peer uses to identifie itself
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// New create a new hanshake with pstr standard
func New(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// Serialize  serializes the handshake to a buffer into this form
// <pstr length> <8 bytes to indicate extensions support> <info hash> <peer id> <==== buffer
func (h *Handshake) Serialize() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}

// Read parses a handshake from a stream
func Read(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)

	if err != nil {
		return nil, err
	}

	pstrlen := int(lengthBuf[0])
	if pstrlen == 0 {
		return nil, fmt.Errorf("Disconnecting..")
	}
	handshakeBuff := make([]byte, pstrlen+48)
	_, err = io.ReadFull(r, handshakeBuff)
	if err != nil {
		return nil, err
	}
	var infoHash, peerID [20]byte
	copy(infoHash[:], handshakeBuff[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuff[pstrlen+8+20:])

	h := Handshake{
		Pstr:     string(handshakeBuff[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil
}
