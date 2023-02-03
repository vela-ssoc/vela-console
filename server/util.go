package server

import (
	"bytes"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/vela-ssoc/vela-kit/vela"
	"io/ioutil"
)

func files(line string) []string {
	names := make([]string, 0)
	dir, _ := ioutil.ReadDir(Instance.cfg.Script)
	for _, f := range dir {
		names = append(names, f.Name())
	}
	return names
}

func empty(line string) []string {
	return []string{}
}

func newCompleter(p *prompt) *readline.PrefixCompleter {

	return readline.NewPrefixCompleter(
		readline.PcItem("use", readline.PcItemDynamic(p.useFunc)),
		readline.PcItem("load", readline.PcItemDynamic(files)),
		readline.PcItem("list"),
		readline.PcItem("help"),
		readline.PcItem("quit"),
		readline.PcItem("clear"),
		readline.PcItem("pwd"),
	)
}

func display(prefix string, task *vela.Task) string {
	var buff bytes.Buffer

	task.Range(func(r *vela.Runner) {
		buff.WriteString(fmt.Sprintf(prefix+"名称: %-46s|\n", r.Name))
		buff.WriteString(fmt.Sprintf(prefix+"类型: %-46s|\n", r.Type))
		buff.WriteString(fmt.Sprintf(prefix+"状态: %-46s|\n", r.Status))
		buff.WriteString(fmt.Sprintf(prefix+"属性: %-46v|\n", r.Private))
		buff.WriteString("+----------------------------------------------------+\n")
	})

	return buff.String()
}
