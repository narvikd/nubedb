package config

import (
	"flag"
	"fmt"
)

type Config struct {
	ID               string
	ApiPort          int
	ApiAddress       string
	ConsensusPort    int
	ConsensusAddress string
	GrpcPort         int
	GrpcAddress      string
}

func New() Config {
	id := flag.String("id", "node1", "")
	//host := flag.String("host", "localhost", "")
	apiPort := flag.Int("port", 3001, "")
	//consensusPort := flag.Int("consensus-port", 4001, "")
	//grpcPort := flag.Int("grpc-port", 5001, "")
	flag.Parse()

	consensusPort := *apiPort + 1
	grpcPort := consensusPort + 1

	cfg := Config{
		ID:               *id,
		ApiPort:          *apiPort,
		ApiAddress:       makeAddr(*apiPort),
		ConsensusPort:    consensusPort,
		ConsensusAddress: makeAddr(consensusPort),
		GrpcPort:         grpcPort,
		GrpcAddress:      makeAddr(grpcPort),
	}
	fmt.Println("Config:", cfg)
	return cfg
}

func makeAddr(port int) string {
	const host = "localhost"
	return fmt.Sprintf("%s:%v", host, port)
}
