package torrentfile

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update torrnt json file")

func TestOpen(t *testing.T) {
	torrent, err := Open("../../testdata/archlinux-2019.12.01-x86_64.iso.torrent")
	require.Nil(t, err)

	jsonPath := "../../testdata/archlinux-2019.12.01-x86_64.iso.json"
	if *update {
		serialized, err := json.MarshalIndent(torrent, "", " ")
		require.Nil(t, err)
		ioutil.WriteFile(jsonPath, serialized, 0644)
	}
	expected := TorrentFile{}
	jsonData, err := ioutil.ReadFile(jsonPath)
	require.Nil(t, err)
	err = json.Unmarshal(jsonData, &expected)
	require.Nil(t, err)
	assert.Equal(t, expected, torrent)
}

func TestToTorrent(t *testing.T) {
	tests := map[string]struct {
		input  *bencodeTorrent
		output TorrentFile
		fails  bool
	}{
		"correct conversion": {
			input: &bencodeTorrent{
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo{
					Pieces:       "1234567890abcdefghijabcdefghij1234567890",
					PiecesLength: 262144,
					Length:       351272960,
					Name:         "debian-10.2.0-amd64-netinst.iso",
				},
			},
			output: TorrentFile{
				Announce: "http://bttracker.debian.org:6969/announce",
				InfoHash: [20]byte{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
				PieceHashes: [][20]byte{
					{49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106},
					{97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
				},
				PieceLength: 262144,
				Length:      351272960,
				Name:        "debian-10.2.0-amd64-netinst.iso",
			},
			fails: false,
		},
		"not enough bytes in pieces": {
			input: &bencodeTorrent{
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo{
					Pieces:       "1234567890abcdefghijabcdef", // Only 26 bytes
					PiecesLength: 262144,
					Length:       351272960,
					Name:         "debian-10.2.0-amd64-netinst.iso",
				},
			},
			output: TorrentFile{},
			fails:  true,
		},
	}
	for _, test := range tests {
		to, err := test.input.toTorrentFile()
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, to)
	}
}
