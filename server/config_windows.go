package server

const format string = "%s »"

type config struct {
	Enable  bool   `lua:"enable,true"`
	Network string `lua:"network"`
	Address string `lua:"address"`
	Script  string `lua:"script"`
}

func normal() *config {
	return &config{Enable: true, Network: "udp", Address: "127.0.0.1:13389", Script: "."}
}
