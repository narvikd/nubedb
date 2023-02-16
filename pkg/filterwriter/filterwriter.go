package filterwriter

import (
	"io"
	"strings"
)

type Writer struct {
	innerWriter io.Writer
	filters     []string
}

func (fw *Writer) Write(input []byte) (n int, err error) {
	for _, filter := range fw.filters {
		if strings.Contains(string(input), filter) {
			return len(input), nil
		}
	}
	return fw.innerWriter.Write(input)
}

func New(innerWriter io.Writer, filters []string) *Writer {
	return &Writer{
		innerWriter: innerWriter,
		filters:     filters,
	}
}
