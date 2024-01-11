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
	pstrLen := len(h.Pstr)
	bufLen := pstrLen + 49
	buf := make([]byte, bufLen)
	buf[0] = byte(pstrLen)
	copy(buf[1:], []byte(h.Pstr))
	// 8 reserved bytes to indicate extension support !
	copy(buf[1+pstrLen+8:], h.InfoHash[:])
	copy(buf[1+pstrLen+8+20:], h.PeerID[:])
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
		return nil, fmt.Errorf("Pstr Length cannot be 0")
	}
	handshakeBuff := make([]byte, pstrlen+49)
	_, err = io.ReadFull(r, handshakeBuff)
	if err != nil {
		return nil, err
	}
	var infoHash, peerID [20]byte
	copy(infoHash[:], handshakeBuff[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuff[1+pstrlen+8+20:])

	h := &Handshake{
		Pstr:     string(handshakeBuff[1:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}
	return h, nil
}
