package main

import (
	"flag"
	"fmt"
)

type FlagValues []string

const Version = "0.0.6"

var ListenPort string
var Files FlagValues
var Action string

func (fv *FlagValues) Set(value string) error {
	*fv = append(*fv, value)
	return nil
}

func (i *FlagValues) String() string {
	return "-"
}

func main() {
	port := flag.Int("p", 8888, "listening port")
	flag.Var(&Files, "f", "file to watch (mutliple -f accepted, default = current dir)")
	conditional := flag.Bool("c", false, "display a conditional script sample")
	action := flag.String("a", "location.reload()", "custom action")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(Version)
		return
	}

	if len(Files) == 0 {
		Files = append(Files, ".")
	}

	ListenPort = fmt.Sprintf("localhost:%d", *port)
	Action = *action

	fmt.Println("Watching", Files)
	fmt.Println("Add the following line to your HTML page:")
	fmt.Println("--")
	if *conditional {
		fmt.Println(`<script>
  if (location.hostname === 'localhost') document.write('<script src="http://localhost:8888/refresh"></' + 'script>')
</script>`)
	} else {
		fmt.Printf("<script src=\"http://%s/refresh\"></script>\n", ListenPort)
	}
	fmt.Println("--")
	StartServer()
}
