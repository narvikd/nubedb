package config

import (
	"flag"
	"fmt"
)

type Config struct {
	ID               string
	ApiPort          int
	ConsensusHost    string
	ConsensusPort    int
	ConsensusAddress string
}

func New() Config {
	id := flag.String("id", "node_1", "")
	apiPort := flag.Int("api-port", 3001, "")
	consensusHost := flag.String("consensus-host", "localhost", "")
	consensusPort := flag.Int("consensus-port", 4001, "")
	flag.Parse()

	consensusAddress := fmt.Sprintf("%s:%v", *consensusHost, *consensusPort)
	return Config{
		ID:               *id,
		ApiPort:          *apiPort,
		ConsensusHost:    *consensusHost,
		ConsensusPort:    *consensusPort,
		ConsensusAddress: consensusAddress,
	}
}
