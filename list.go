package thyme

import (
	"fmt"
	"io"
)

func List(w io.Writer, stream *Stream) {
	fmt.Fprintf(w, "%s", stream.Print())
}
