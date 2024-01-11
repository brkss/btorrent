package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/brkss/btorrent/src/p2p"
	"github.com/jackpal/bencode-go"
)

const PORT uint16 = 6881

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces       string `bencode:"pieces"`
	PiecesLength int    `bencode:"piece length"`
	Length       int    `bencode:"length"`
	Name         string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

func (t *TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return nil
	}

	peers, err := t.requestPeers(peerID, PORT)
	if err != nil {
		return err
	}

	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}

	buf, err := torrent.Download()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()
	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}

	return bto.toTorrentFile()
}

func (b *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *b)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (b *bencodeInfo) splitPieceHash() ([][20]byte, error) {
	hashlen := 20
	buf := []byte(b.Pieces)
	if len(buf)%hashlen != 0 {
		err := fmt.Errorf("Recieved malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) % hashlen
	hashes := make([][20]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[(i*hashlen):(i+1)*hashlen])
	}
	return hashes, nil
}

func (b *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := b.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}

	pieceHash, err := b.Info.splitPieceHash()
	if err != nil {
		return TorrentFile{}, nil
	}

	t := TorrentFile{
		Announce:    b.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHash,
		PieceLength: b.Info.PiecesLength,
		Length:      b.Info.Length,
		Name:        b.Info.Name,
	}
	return t, nil
}
