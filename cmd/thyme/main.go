package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/sourcegraph/thyme"
)

var CLI = flags.NewNamedParser("thyme", flags.PrintErrors|flags.PassDoubleDash)

func init() {
	CLI.Usage = `
thyme - automatically track which applications you use and for how long.

  \|//   thyme is a simple time tracker that tracks active window names and collects
 W Y/    statistics over active, open, and visible windows. Statistics are collected
  \|  ,  into a local JSON file, which is used to generate a pretty HTML report.
   \_/
    \
     \_  thyme is a local CLI tool and does not send any data over the network.

Example usage:

  thyme dep
  thyme track -o <file>
  thyme show  -i <file> -w stats > viz.html

`

	if _, err := CLI.AddCommand("track", "record current windows", "Record current window metadata as JSON printed to stdout or a file. If a filename is specified and the file already exists, Thyme will append the new snapshot data to the existing data.", &trackCmd); err != nil {
		log.Fatal(err)
	}
	if _, err := CLI.AddCommand("show", "visualize data", "Generate an HTML page visualizing the data from a file written to by `thyme track`.", &showCmd); err != nil {
		log.Fatal(err)
	}
	if _, err := CLI.AddCommand("dep", "dep install instructions", "Show installation instructions for required external dependencies (which vary depending on your OS and windowing system).", &depCmd); err != nil {
		log.Fatal(err)
	}
	if _, err := CLI.AddCommand("watch", "", "record current windows on an interval", &watchCmd); err != nil {
		log.Fatal(err)
	}
	if _, err := CLI.AddGroup("shortcuts", "shortcuts for frequently used commands", &shortcutOptions); err != nil {
		log.Fatal(err)
	}
	CLI.SubcommandsOptional = true
}

type ShortcutOptions struct {
	Watch bool `long:"watch" short:"w" description:"sets up default watch"`
	Show  bool `long:"show" short:"s" description:"sets up default show"`
}

var shortcutOptions ShortcutOptions

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

// WatchCmd is the subcommand that tracks application usage on an interval
type WatchCmd struct {
	Out      string `long:"out" short:"o" description:"output file" default:"thyme.json"`
	Interval uint32 `long:"interval" short:"n" description:"seconds between records" default:"30"`
}

var watchCmd WatchCmd

func (c *WatchCmd) Execute(args []string) error {
	trackCmd.Out = c.Out
	err := trackCmd.Execute(args)
	if err != nil {
		return err
	}
	for range time.Tick(time.Duration(c.Interval) * time.Second) {
		err := trackCmd.Execute(args)
		if err != nil {
			return err
		}
	}
	return nil
}

// ShowCmd is the subcommand that reads the data emitted by the track
// subcommand and displays the data to the user.
type ShowCmd struct {
	In    string `long:"in" short:"i" description:"input file"`
	What  string `long:"what" short:"w" description:"what to show {list,stats}" default:"list"`
	Serve bool   `long:"serve" short:"s" description:"serves content via http"`
}

var showCmd ShowCmd

func (c *ShowCmd) Execute(args []string) error {
	if c.Serve {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open(c.In)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
			defer f.Close()

			var stream thyme.Stream
			if err := json.NewDecoder(f).Decode(&stream); err != nil {
				fmt.Fprint(w, err)
				return
			}
			thyme.Stats(w, &stream)
		})
		http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open(c.In)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
			defer f.Close()

			var stream thyme.Stream
			if err := json.NewDecoder(f).Decode(&stream); err != nil {
				fmt.Fprint(w, err)
				return
			}
			thyme.List(w, &stream)
		})
		fmt.Println("Starting server...")
		http.ListenAndServe(":6090", nil)
	}
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
			if err := thyme.Stats(os.Stdout, &stream); err != nil {
				return err
			}
		case "list":
			fallthrough
		default:
			thyme.List(os.Stdout, &stream)
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
			if _, isFlagsErr := err.(*flags.Error); isFlagsErr {
				CLI.WriteHelp(os.Stderr)
				return nil
			} else {
				return err
			}
		}
		showOpt := CLI.FindOptionByLongName("show")
		if showOpt != nil {
			showCmd.In = watchCmd.Out
			showCmd.What = "show"
			showCmd.Serve = true
			go showCmd.Execute([]string{})
		}
		opt := CLI.FindOptionByLongName("watch")
		if opt != nil {
			watchCmd.Execute([]string{})
		}
		return nil
	}

	if err := run(); err != nil {
		os.Exit(1)
	}
}

func getTracker() (thyme.Tracker, error) {
	switch runtime.GOOS {
	case "windows":
		return thyme.NewTracker("windows"), nil
	case "darwin":
		return thyme.NewTracker("darwin"), nil
	default:
		return thyme.NewTracker("linux"), nil
	}
}
