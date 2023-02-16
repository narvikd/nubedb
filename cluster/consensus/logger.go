package consensus

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"nubedb/pkg/filterwriter"
	"os"
)

// newConsensusFilterWriter returns a filter-writer which suppresses raft's errors that had been debug errors instead
func newConsensusFilterWriter() *filterwriter.Writer {
	const errCannotSnapshotNow = "cannot restore snapshot now, wait until the configuration entry at"
	filters := []string{raft.ErrNothingNewToSnapshot.Error(), errCannotSnapshotNow}
	return filterwriter.New(os.Stderr, filters)
}

// TODO: Maybe this should be decoupled
func (n *Node) setConsensusLogger(cfg *raft.Config) {
	fw := newConsensusFilterWriter()
	l := hclog.New(&hclog.LoggerOptions{
		Name:   "consensus",
		Level:  hclog.LevelFromString("DEBUG"),
		Output: fw,
	})

	n.consensusLogger = l

	cfg.LogOutput = fw
	cfg.Logger = l
}

func (n *Node) LogWrapErr(err error, message string) {
	n.consensusLogger.Error(fmt.Errorf("%s: %w", message, err).Error())
}
