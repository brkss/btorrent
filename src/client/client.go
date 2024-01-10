package client

import (
	"net"

	"github.com/brkss/btorrent/src/bitfield"
	"github.com/brkss/btorrent/src/peer"
)

// client is a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peer.Peer
	infoHash [20]byte
	peerID   [20]byte
}
