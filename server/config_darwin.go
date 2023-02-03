package server

const format string = "\033[31m%s»\033[0m "

type config struct {
	Enable  bool   `lua:"enable,false"`
	Network string `lua:"network"`
	Address string `lua:"address"`
	Script  string `lua:"script"`
}

func normal() *config {
	return &config{Enable: true, Network: "unix", Address: "rock.sock", Script: "."}
}
