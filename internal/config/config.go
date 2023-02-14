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
	ConsensusPort    int
	ConsensusAddress string
	GrpcPort         int
	GrpcAddress      string
}

func New() Config {
	id := flag.String("id", "node_1", "")
	host := flag.String("host", "localhost", "")
	apiPort := flag.Int("api-port", 3001, "")
	consensusPort := flag.Int("consensus-port", 4001, "")
	grpcPort := flag.Int("grpc-port", 5001, "")
	flag.Parse()

	return Config{
		ID:               *id,
		Host:             *host,
		ApiPort:          *apiPort,
		ApiAddress:       makeAddr(*host, *apiPort),
		ConsensusPort:    *consensusPort,
		ConsensusAddress: makeAddr(*host, *consensusPort),
		GrpcPort:         *grpcPort,
		GrpcAddress:      makeAddr(*host, *grpcPort),
	}
}

func makeAddr(host string, port int) string {
	return fmt.Sprintf("%s:%v", host, port)
}
