package client

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"os"
	"runtime"
)

var url string

// unix://
func init() {
	if runtime.GOOS == "windows" {
		url = "tcp://127.0.0.1:3399"
	} else {
		url = "unix://vela.sock"
	}
}

func Do() {
	var err error
	var s auxlib.URL

	if len(os.Args) <= 1 {
		s, err = auxlib.NewURL(url)
	} else {
		argv := os.Args[1]
		if argv == "-h" || argv == "--help" {
			fmt.Println("vela-cli unix://rock.sock  如果不添加参数 默认 win: udp://127.0.0.1:3399 linux: unix://vela.sock")
			return
		}
		s, err = auxlib.NewURL(argv)
	}

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = readline.DialRemote(s.Scheme(), s.Host())
	if err != nil {
		fmt.Println(err.Error())
	}
}
