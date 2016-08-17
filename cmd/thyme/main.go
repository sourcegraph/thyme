package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/kovetskiy/godocs"
	"github.com/sourcegraph/thyme"
)

var version = "[manual build]"

var usage = `thyme - automatically track which applications you use and for how long.

  \|//   thyme is simple time tracker, which tracks active window names and
 W Y/    collect statistics over active, open and visible windows. Statistics
  \|  ,  is collected into local JSON file, which can be transformed
   \_/   into pretty HTML report.
    \
     \_  thyme is a local CLI tool and do not send any info over network.

Usage:
  thyme track -o <file>
  thyme show  -i <file> -w <what>
  thyme dep
  thyme -h | --help

Commands:
  track               Will track current windows state into new record inside
                       specified JSON file.

  show                Will show computed applications usage based on collected
                       statistics.

  dep                 Show miscallaneous installation info.

Options:
  -h --help           Show this help.
  -v --version        Show program version. Will be equal to '[manual build]' if
                       tool is built via go get without package manager.

Track options:
  -o --output <file>  Specify JSON file to collect statistics into.

Show options:
  -i --input <file>   Specify JSON file to read statistic into.
  -w --what <what>    Display specified statistics. Can be either
                       'list' (will output plaintext) or
                       'stats' (will output HTML).
                       [default: list]
`

// TrackCmd is the subcommand that tracks application usage.
type TrackCmd struct {
	Out string
}

func (c TrackCmd) Execute() error {
	t, err := getTracker()
	if err != nil {
		return err
	}
	snap, err := t.Snap()
	if err != nil {
		return err
	}

	if c.Out == "" {
		out, err := json.MarshalIndent(snap, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	} else {
		var stream thyme.Stream
		if _, err := os.Stat(c.Out); err == nil {
			if err := func() error {
				f, err := os.Open(c.Out)
				if err != nil {
					return err
				}
				defer f.Close()
				if err := json.NewDecoder(f).Decode(&stream); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		stream.Snapshots = append(stream.Snapshots, snap)
		f, err := os.Create(c.Out)
		if err != nil {
			return err
		}
		if err := json.NewEncoder(f).Encode(stream); err != nil {
			return err
		}
	}

	return nil
}

// ShowCmd is the subcommand that reads the data emitted by the track
// subcommand and displays the data to the user.
type ShowCmd struct {
	In   string
	What string
}

func (c ShowCmd) Execute() error {
	if c.In == "" {
		var snap thyme.Snapshot
		if err := json.NewDecoder(os.Stdin).Decode(&snap); err != nil {
			return err
		}
		for _, w := range snap.Windows {
			fmt.Printf("%+v\n", w.Info())
		}
	} else {
		var stream thyme.Stream
		f, err := os.Open(c.In)
		if err != nil {
			return err
		}
		defer f.Close()

		if err := json.NewDecoder(f).Decode(&stream); err != nil {
			return err
		}
		switch c.What {
		case "stats":
			if err := thyme.Stats(&stream); err != nil {
				return err
			}
		case "list":
			fallthrough
		default:
			thyme.List(&stream)
		}
	}
	return nil
}

type DepCmd struct{}

func (c DepCmd) Execute() error {
	t, err := getTracker()
	if err != nil {
		return err
	}
	fmt.Println(t.Deps())
	return nil
}

func main() {
	args, err := godocs.Parse(usage, "thyme "+version, godocs.UsePager)
	if err != nil {
		panic(err)
	}

	switch {
	case args["track"].(bool):
		err = TrackCmd{
			Out: args["--output"].(string),
		}.Execute()

	case args["show"].(bool):
		err = ShowCmd{
			In:   args["--input"].(string),
			What: args["--what"].(string),
		}.Execute()

	case args["dep"].(bool):
		err = DepCmd{}.Execute()
	}

	if err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}

func getTracker() (thyme.Tracker, error) {
	switch runtime.GOOS {
	case "windows":
		return nil, fmt.Errorf("Windows is unsupported")
	case "darwin":
		return thyme.NewDarwinTracker(), nil
	default:
		return thyme.NewLinuxTracker(), nil
	}
}
