package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"os"
	"github.com/fhs/gompd/mpd"
)

const (
	Seperator     = " — "
	UnknownArtist = "Unknown Artist"
	UnknownAlbum  = "Unknown Album"
)

func main() {
	println("test")
	client, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var lines []string
	Tracks, err := client.ListAllInfo("/")
	if err != nil {
		panic(err)
	}

	var AlbumsFound []string
	var ArtistsFound []string
	for _, Track := range Tracks {
		var Artist string
		if len(Track["AlbumArtist"]) == 0 {
			if len(Track["Artist"]) == 0 {
				Artist = UnknownArtist
				ArtistsFound = append(ArtistsFound, Artist)
			} else {
				Artist = Track["Artist"]
			}
		} else {
			Artist = Track["AlbumArtist"]
		}

		if !contains(ArtistsFound, Artist) {
			lines = append(lines, fmt.Sprintf("%s", Artist))
			ArtistsFound = append(ArtistsFound, Artist)
		}

		var Album string
		if len(Track["Album"]) != 0 {
			Album = Track["Album"]
		} else {
			Album = UnknownAlbum
			AlbumsFound = append(AlbumsFound, Album)
		}

		if !contains(AlbumsFound, Album) {
			lines = append(lines, fmt.Sprintf("%s%s%s", Artist, Seperator, Album))
			AlbumsFound = append(AlbumsFound, Album)
		}

		if len(Track["Title"]) != 0 {
			lines = append(lines, fmt.Sprintf("%s%s%s%s%s", Artist, Seperator, Album, Seperator, Track["Title"]))
		}
	}

	cmd := exec.Command("wofi", "--show", "dmenu", "-p", "Name", "-i")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(stdout)
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	for _, line := range lines {
		io.WriteString(stdin, line+"\n")
	}
	stdin.Close()

	var ID *mpd.PromisedId
	Added := false

	CommandList := client.BeginCommandList()
	for scanner.Scan() {
		Selection := strings.Split(scanner.Text(), Seperator)
		CachedLength := len(Selection)
		if CachedLength > 4 {
			return
		}

		for _, Track := range Tracks {
			_, HasAlbumArtist := Track["AlbumArtist"]
			_, HasArtist := Track["Artist"]
			ArtistUnknown := !HasAlbumArtist && !HasArtist

			ArtistMatch :=
				(Selection[0] == Track["AlbumArtist"] ||
					Selection[0] == Track["Artist"]) ||
					(Selection[0] == UnknownArtist &&
						ArtistUnknown)

			if !ArtistMatch {
				continue
			}

			if CachedLength > 1 {
				if Selection[1] != Track["Album"] &&
					Selection[1] != UnknownAlbum {
					continue
				}
				if CachedLength > 2 {
					if Selection[2] != Track["Title"] {
						continue
					}
				}
			}

			if Added {
				CommandList.AddId(Track["file"], -1)
			} else {
				Added = true
				ID = CommandList.AddId(Track["file"], -1)
			}
		}
	}

	if Added {
		CommandList.End()
		id, _ := ID.Value()
		client.PlayId(id)
		client.Pause(false)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
