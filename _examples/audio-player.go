package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mathom/gumble/gumble"
	"github.com/mathom/gumble/gumble_ffmpeg"
	"github.com/mathom/gumble/gumbleutil"
)

func main() {
	files := make(map[string]string)
	var stream *gumble_ffmpeg.Stream

	gumbleutil.Main(func(client *gumble.Client) {
		var err error
		stream, err = gumble_ffmpeg.New(client)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}

		client.Attach(gumbleutil.AutoBitrate)

		for _, file := range flag.Args() {
			key := filepath.Base(file)
			files[key] = file
		}
	}, gumbleutil.Listener{
		// Connect event
		Connect: func(e *gumble.ConnectEvent) {
			fmt.Printf("audio player loaded! (%d files)\n", len(files))
		},

		// Text message event
		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}
			file, ok := files[e.Message]
			if !ok {
				return
			}
			if err := stream.Play(file); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Playing %s\n", file)
			}
		},
	})
}
