package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

//77.174.62.158:53165

func genTorrentHash() [20]byte {
	hash := "9f3e5fc62e9d0e80ee322a3555896ed6ef498985"
	hexVal, err := hex.DecodeString(hash)
	if err != nil {
		log.Fatal("got error converting hash : ", err)
	}
	var hashInfo [20]byte
	copy(hashInfo[:], hexVal[:])
	return hashInfo
}

func genHandshakeMessage() []byte {
	pstr := "BitTorrent protocol"
	//infoHash := []byte("9f3e5fc62e9d0e80ee322a3555896ed6ef498985")
	infoHash := genTorrentHash()
	var peerID [20]byte

	fmt.Println("hash length : ", len(infoHash), infoHash[0])

	_, err := rand.Read(peerID[:])
	if err != nil {
		log.Fatal("got error : ", err)
	}

	msg := make([]byte, len(pstr)+49)
	msg[0] = byte(len(pstr))
	curr := 1
	curr += copy(msg[curr:], []byte(pstr))
	curr += copy(msg[curr:], make([]byte, 8))
	curr += copy(msg[curr:], infoHash[:])
	curr += copy(msg[curr:], peerID[:])
	return msg
}

func main() {

	handshake := genHandshakeMessage()

	con, err := net.DialTimeout("tcp", "80.78.21.211:53463", time.Second*3)
	defer con.SetDeadline(time.Time{})
	if err != nil {
		log.Fatal("got err :", err)
	}

	n, err := con.Write(handshake)
	fmt.Println("written bytes : ", n)
	if err != nil {
		log.Fatal("failed to write to connection : ", err)
	}
	buff := make([]byte, 1)
	n, err = io.ReadFull(con, buff)
	fmt.Println("readed bytes : ", n, buff)
	if err != nil {
		log.Fatal("failed reading from connection : ", err)
	}

	defer con.Close()

	log.Print("connection : ", con)

}
