package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/jessevdk/go-flags"
	"github.com/sourcegraph/thyme"
)

var CLI = flags.NewNamedParser("thyme", flags.PrintErrors|flags.PassDoubleDash)

func init() {
	if _, err := CLI.AddCommand("track", "", "record current windows", &trackCmd); err != nil {
		log.Fatal(err)
	}
	if _, err := CLI.AddCommand("show", "", "visualize data", &showCmd); err != nil {
		log.Fatal(err)
	}
	if _, err := CLI.AddCommand("dep", "", "external dependencies that need to be installed", &depCmd); err != nil {
		log.Fatal(err)
	}
}

// TrackCmd is the subcommand that tracks application usage.
type TrackCmd struct {
	Out string `long:"out" short:"o" description:"output file"`
}

var trackCmd TrackCmd

func (c *TrackCmd) Execute(args []string) error {
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
	In   string `long:"in" short:"i" description:"input file"`
	What string `long:"what" short:"w" description:"what to show {list,stats}" default:"list"`
}

var showCmd ShowCmd

func (c *ShowCmd) Execute(args []string) error {
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

var depCmd DepCmd

func (c *DepCmd) Execute(args []string) error {
	t, err := getTracker()
	if err != nil {
		return err
	}
	fmt.Println(t.Deps())
	return nil
}

func main() {
	run := func() error {
		_, err := CLI.Parse()
		if err != nil {
			return err
		}
		return nil
	}

	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func getTracker() (thyme.Tracker, error) {
	switch runtime.GOOS {
	case "windows":
		return thyme.NewWindowsTracker(), nil
	case "darwin":
		return thyme.NewDarwinTracker(), nil
	default:
		return thyme.NewLinuxTracker(), nil
	}
}
