package consensus

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"io"
	"os"
	"strings"
)

type filterWriter struct {
	innerWriter io.Writer
	filter      string
}

func (fw *filterWriter) Write(input []byte) (n int, err error) {
	if strings.Contains(string(input), fw.filter) {
		return len(input), nil
	}
	return fw.innerWriter.Write(input)
}

func newFilterWriter(innerWriter io.Writer, filter string) *filterWriter {
	return &filterWriter{
		innerWriter: innerWriter,
		filter:      filter,
	}
}

func setConsensusLogger(cfg *raft.Config) {
	// Suppresses "nothing new to snapshot" errors
	fw := newFilterWriter(os.Stderr, raft.ErrNothingNewToSnapshot.Error())

	cfg.LogOutput = fw
	cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Name:   "consensus",
		Level:  hclog.LevelFromString("DEBUG"),
		Output: fw,
	})
}
