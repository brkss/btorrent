package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

const (
	// MsgChoke chokes the receiver
	MsgChoke messageID = 0
	// MsgUnchoke unchokes the receiver
	MsgUnchoke messageID = 1
	// MsgInterested expresses interest in receiving data
	MsgInterested messageID = 2
	// MsgNotInterested expresses disinterest in receiving data
	MsgNotInterested messageID = 3
	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgHave messageID = 4
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield messageID = 5
	// MsgRequest requests a block of data from the receiver
	MsgRequest messageID = 6
	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece messageID = 7
	// MsgCancel cancels a request
	MsgCancel messageID = 8
)

type Message struct {
	ID      messageID
	Payload []byte
}

// FormatRequest create a request message
func FormatRequest(length int, index int, begin int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

// FormatHave create a Have message
func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

// ParsePiece parses a piece message and copy its content into a buffer
func ParsePiece(index int, buf []byte, message *Message) (int, error) {
	if message.ID != MsgPiece {
		return 0, fmt.Errorf("Expected Message Piece (ID: %d). got ID %d", MsgPiece, message.ID)
	}
	if len(message.Payload) < 8 {
		return 0, fmt.Errorf("Payload too short %d < 8", len(message.Payload))
	}
	parsedIndex := int(binary.BigEndian.Uint16(message.Payload[0:4]))
	if index != parsedIndex {
		return 0, fmt.Errorf("Expected index %d got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint16(message.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin offset too high : %d >= %d", begin, len(buf))
	}
	data := message.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("Data too long [%d] for buffer [%d]", begin+len(data), len(buf))
	}
	copy(buf[begin:], data)
	return len(data), nil
}

// ParseHave parses a have message
func ParseHave(msg *Message) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("Expected Have Message (%d) got (%d)", MsgHave, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("Expected Payload with 4 as length got %d", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interepets nil as Keep-Alive message

func (message *Message) Serialize() []byte {
	if message == nil {
		return make([]byte, 4)
	}
	len := uint32(len(message.Payload) + 1) // +1 for message ID
	buff := make([]byte, len+4)
	binary.BigEndian.PutUint32(buff[0:4], uint32(len))
	buff[4] = byte(message.ID)
	copy(buff[5:], message.Payload)
	return buff
}

// Read parses a message from a stream return nil on keep-alive message
func Read(r io.Reader) (*Message, error) {
	lengthBuff := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuff)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuff)

	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	m := &Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}
	return m, nil
}

func (m *Message) name() string {
	if m == nil {
		return "Keep-Alive"
	}
	switch m.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "NotInterested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown#%d", m.ID)
	}
}

func (m *Message) String() string {
	if m == nil {
		return m.name()
	}
	return fmt.Sprintf("%s [%d]", m.name(), len(m.Payload))
}
