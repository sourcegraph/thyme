package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/beyang/thyme"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	t := thyme.NewLinuxTracker()
	snap, err := t.Snap()
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))

	for _, w := range snap.Windows {
		fmt.Printf("%+v\n", w.Info())
	}
	return nil
}
