package thyme

import "fmt"

func List(stream *Stream) {
	fmt.Printf("%s", stream.Print())
}
