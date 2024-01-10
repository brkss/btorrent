package p2p

import (
	"github.com/brkss/btorrent/src/peer"
)

const (
	MAX_BLOCK_SIZE = 16384 // max number of bits a request can ask for
	MAX_BLOCK_LOG  = 5     // max number of unfulfilled request client have in its pipeline
)

// hold data required to download a torrent from a list of peers
type Torrent struct {
	Peers       []peer.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceProgress struct {
	index int
}
