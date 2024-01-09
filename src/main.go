package main

import (
	"fmt"
	"log"
	"os"

	"github.com/brkss/btorrent/src/torrentfile"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Invalid Argements")
		return
	}
	torrentPath := os.Args[1]
	//output := os.Args[2];
	_, err := torrentfile.Open(torrentPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Invalid Torrent File : %s\n %s", torrentPath, err))
	}

}
