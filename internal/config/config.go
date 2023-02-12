package config

import (
	"flag"
	"fmt"
)

type Config struct {
	ID               string
	Host             string
	ApiPort          int
	ApiAddress       string
	ConsensusHost    string
	ConsensusPort    int
	ConsensusAddress string
}

func New() Config {
	id := flag.String("id", "node_1", "")
	host := flag.String("host", "localhost", "")
	apiPort := flag.Int("api-port", 3001, "")
	consensusPort := flag.Int("consensus-port", 4001, "")
	flag.Parse()

	apiAddress := fmt.Sprintf("%s:%v", *host, *apiPort)
	consensusAddress := fmt.Sprintf("%s:%v", *host, *consensusPort)
	return Config{
		ID:               *id,
		Host:             *host,
		ApiPort:          *apiPort,
		ApiAddress:       apiAddress,
		ConsensusPort:    *consensusPort,
		ConsensusAddress: consensusAddress,
	}
}
