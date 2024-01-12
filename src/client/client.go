package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/brkss/btorrent/src/bitfield"
	"github.com/brkss/btorrent/src/handshake"
	"github.com/brkss/btorrent/src/message"
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

func completeHandshake(conn net.Conn, infoHash [20]byte, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	req := handshake.New(infoHash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("Invalid Info Hash Expected %x and got %x", infoHash, res.InfoHash)
	}
	return res, nil
}

func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(time.Second * 5))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		return nil, fmt.Errorf("expected a message bitfield but got : %d", msg.ID)
	}

	return msg.Payload, nil
}

// New create new connection with client, send a handshake and reciece a handshake
// return error if any of those fail!
func New(peer peer.Peer, peerID, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), time.Second*3)
	if err != nil {

		return nil, err
	}
	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		return nil, err
	}

	bitfield, err := recvBitfield(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   false,
		Bitfield: bitfield,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil

}

// Read reads and consumes a message from the connection
func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

// SendRequest send requect to client in the current connection
func (c *Client) SendReuqest(index, begin, length int) error {
	req := message.FormatRequest(length, index, begin)
	_, err := c.Conn.Write(req.Serialize())

	return err
}

// SendIntrested sends an Intresseted message to a peer
func (c *Client) SendIntrested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendNotIntrested sends an not intrested message to a peer
func (c *Client) SendNotIntrested() error {
	msg := message.Message{ID: message.MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendUnchoke sends an Unchoke message to a peer
func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())

	return err
}

// SendHave sends a Have message to client
func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())

	return err
}
