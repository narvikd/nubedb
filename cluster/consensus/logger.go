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
	filters     []string
}

func (fw *filterWriter) Write(input []byte) (n int, err error) {
	for _, filter := range fw.filters {
		if strings.Contains(string(input), filter) {
			return len(input), nil
		}
	}
	return fw.innerWriter.Write(input)
}

func newFilterWriter(innerWriter io.Writer, filters []string) *filterWriter {
	return &filterWriter{
		innerWriter: innerWriter,
		filters:     filters,
	}
}

// Suppresses raft's errors that should be debug errors
func newConsensusFilterWriter() *filterWriter {
	const errCannotSnapshotNow = "cannot restore snapshot now, wait until the configuration entry at"
	filters := []string{raft.ErrNothingNewToSnapshot.Error(), errCannotSnapshotNow}
	return newFilterWriter(os.Stderr, filters)
}

func setConsensusLogger(cfg *raft.Config) {
	fw := newConsensusFilterWriter()
	cfg.LogOutput = fw
	cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Name:   "consensus",
		Level:  hclog.LevelFromString("DEBUG"),
		Output: fw,
	})
}
