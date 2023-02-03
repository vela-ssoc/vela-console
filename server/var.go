package server

var (
	Instance *Serv
)

func init() {
	Instance = newServ()
}
