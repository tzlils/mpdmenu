package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fhs/gompd/mpd"
)

const (
	Seperator = " — "
)

func main() {
	client, err := mpd.Dial("tcp", "localhost:6600")
	orPanic(err)
	defer client.Close()

	var lines []string
	Tracks, err := client.ListAllInfo("/")
	orPanic(err)

	var AlbumsFound []string
	var ArtistsFound []string
	for _, Track := range Tracks {
		if len(Track["Artist"]) == 0 {
			continue
		}

		if !contains(ArtistsFound, Track["AlbumArtist"]) {
			lines = append(lines, fmt.Sprintf("%s", Track["AlbumArtist"]))
			ArtistsFound = append(ArtistsFound, Track["AlbumArtist"])
		}

		if !contains(AlbumsFound, Track["Album"]) {
			lines = append(lines, fmt.Sprintf("%s%s%s", Track["AlbumArtist"], Seperator, Track["Album"]))
			AlbumsFound = append(AlbumsFound, Track["Album"])
		}

		lines = append(lines, fmt.Sprintf("%s%s%s%s%s", Track["AlbumArtist"], Seperator, Track["Album"], Seperator, Track["Title"]))
	}

	subProcess := exec.Command("rofi", "-dmenu", "-p", "Name", "-i")
	stdin, _ := subProcess.StdinPipe()
	stdout, _ := subProcess.StdoutPipe()
	scanner := bufio.NewScanner(stdout)
	subProcess.Start()
	for _, line := range lines {
		io.WriteString(stdin, line+"\n")
	}
	stdin.Close()

	var ID *mpd.PromisedId
	var Added bool

	CommandList := client.BeginCommandList()
	for scanner.Scan() {
		Selection := strings.Split(scanner.Text(), Seperator)
		if len(Selection) == 0 {
			// No selection
			os.Exit(0)
		} else if len(Selection) == 1 {
			// Artist selection
			for _, Track := range Tracks {
				if Track["AlbumArtist"] == Selection[0] {
					if Added {
						CommandList.AddId(Track["file"], -1)
					} else {
						Added = true
						ID = CommandList.AddId(Track["file"], -1)
					}

				}
			}
		} else if len(Selection) == 2 {
			// Album selection
			for _, Track := range Tracks {
				if Track["AlbumArtist"] == Selection[0] &&
					Track["Album"] == Selection[1] {
					if Added {
						CommandList.AddId(Track["file"], -1)
					} else {
						ID = CommandList.AddId(Track["file"], -1)
						Added = true
					}
				}
			}
		} else if len(Selection) == 3 {
			// Track selection
			for _, Track := range Tracks {
				if Track["AlbumArtist"] == Selection[0] &&
					Track["Album"] == Selection[1] &&
					Track["Title"] == Selection[2] {
					ID = CommandList.AddId(Track["file"], -1)
				}
			}
		} else {
			// ???
			os.Exit(1)
		}
		CommandList.End()
	}

	id, err := ID.Value()
	orPanic(err)
	client.PlayId(id)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func uniq(in []string) (out []string) {
	for _, e := range in {
		if !contains(out, e) {
			out = append(out, e)
		}
	}
	return
}

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}
