package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/brkss/btorrent/src/client"
	"github.com/brkss/btorrent/src/message"
	"github.com/brkss/btorrent/src/peer"
)

const (
	MAX_BACKLOG_SIZE = 16384 // max number of bits a request can ask for
	MAX_BACKLOG      = 5     // max number of unfulfilled request client have in its pipeline
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

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil // as keep-alive message !
	}

	switch msg.ID {
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return nil
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func attemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	// setting deadline help get unresponding client unstuck
	// 30 second is more than enough time to download 262kb piece
	c.Conn.SetDeadline(time.Now().Add(time.Second * 30))
	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		// if the client is unchoked keep sending requests till we have enough unfulffiled requests
		if !state.client.Choked {
			for state.backlog < MAX_BACKLOG && state.requested < pw.length {
				blockSize := MAX_BACKLOG_SIZE
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}
				err := c.SendRequest(state.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}
		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}

func checkIntergrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check!", pw.index)
	}
	return nil
}

func (t *Torrent) startDownloaderWorker(peer peer.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := client.New(peer, t.PeerID, t.InfoHash)
	if err != nil {
		//log.Println("err: ", err)
		log.Printf("could not handshake with client %s, Disconnecting... \n", peer.IP)
		return
	}

	defer c.Conn.Close()

	log.Printf("Complete handshake successfuly with client %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw
			continue
		}

		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Printf("Exiting..")
			workQueue <- pw
			return
		}

		err = checkIntergrity(pw, buf)
		if err != nil {
			log.Printf("failed to check piece [%d] integrity \n", pw.index)
			workQueue <- pw
			continue
		}
		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}
}

func (t *Torrent) calculateBoundsForPeice(index int) (begin, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.PieceLength {
		end = t.PieceLength
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPeice(index)
	return end - begin
}

// Downloads downloads the torrent , it store the whole file in memory !
func (t *Torrent) Download() ([]byte, error) {
	log.Printf("Start downloading: %s\n", t.Name)

	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	result := make(chan *pieceResult)

	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	//fmt.Println("iterating on peers count : ", len(t.Peers))
	// run threads to start downloading torrent
	for _, peer := range t.Peers {
		go t.startDownloaderWorker(peer, workQueue, result)
	}
	fmt.Println("pieces : ", len(t.PieceHashes))
	// collect result into buffer untill full !
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-result
		begin, end := t.calculateBoundsForPeice(res.index)
		copy(buf[begin:end], res.buf[:])
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}

	close(workQueue)
	fmt.Println("returning from Download()?")
	return buf, nil
}
