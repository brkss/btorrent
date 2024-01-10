package client

import (
	"net"

	"github.com/brkss/btorrent/src/bitfield"
)

// client is a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}
