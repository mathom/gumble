package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mathom/gumble/gumble"
)

func main() {
	server := flag.String("server", "localhost:64738", "mumble server address")
	timeout := flag.Int("timeout", 5, "ping timeout (seconds) until failure")
	flag.Parse()

	resp, err := gumble.Ping(*server, time.Second*time.Duration(*timeout))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
	major, minor, patch := resp.Version.SemanticVersion()
	fmt.Printf("Address:         %s\n", resp.Address)
	fmt.Printf("Ping:            %s\n", resp.Ping)
	fmt.Printf("Version:         %d.%d.%d\n", major, minor, patch)
	fmt.Printf("Connected Users: %d\n", resp.ConnectedUsers)
	fmt.Printf("Maximum Users:   %d\n", resp.MaximumUsers)
	fmt.Printf("Maximum Bitrate: %d\n", resp.MaximumBitrate)
}
