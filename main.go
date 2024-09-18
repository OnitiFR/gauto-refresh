package main

import (
	"flag"
	"fmt"
	"strings"
)

type FlagValues []string

const Version = "0.0.10"

var ListenPort string
var BaseURL string
var Files FlagValues
var Action string
var Debug bool
var Delay int

func (fv *FlagValues) Set(value string) error {
	*fv = append(*fv, value)
	return nil
}

func (i *FlagValues) String() string {
	return "-"
}

func DebugLog(format string, args ...interface{}) {
	if Debug {
		format = fmt.Sprintf("DEBUG: %s\n", format)
		fmt.Printf(format, args...)
	}
}

func main() {
	port := flag.Int("p", 8888, "listening port")
	base := flag.String("b", "", "base URL for the script (will listen on all interfaces)")
	flag.Var(&Files, "f", "file to watch (mutliple -f accepted, default = current dir)")
	conditional := flag.Bool("c", false, "display a conditional script sample")
	action := flag.String("a", "location.reload()", "custom action")
	version := flag.Bool("v", false, "show version")
	debug := flag.Bool("d", false, "debug mode")
	delay := flag.Int("t", 50, "delay in milliseconds before reload, for double-reload prevention or build/upload time")
	flag.Parse()

	if *version {
		fmt.Println(Version)
		return
	}

	if len(Files) == 0 {
		Files = append(Files, ".")
	}

	ListenPort = fmt.Sprintf("localhost:%d", *port)

	BaseURL = fmt.Sprintf("http://%s", ListenPort)
	if *base != "" {
		ListenPort = fmt.Sprintf(":%d", *port)
		BaseURL = *base
		BaseURL = strings.TrimRight(BaseURL, "/")
	}

	Action = *action
	Debug = *debug
	Delay = *delay

	DebugLog("debug enabled")
	fmt.Println("Watching", Files)
	fmt.Println("Add the following line to your HTML page:")
	fmt.Println("--")
	if *conditional {
		fmt.Println(`<script>
  if (location.hostname === 'localhost') document.write('<script defer src="http://localhost:8888/refresh"></' + 'script>')
</script>`)
	} else {
		fmt.Printf("<script src=\"%s/refresh\"></script>\n", BaseURL)
	}
	fmt.Println("--")
	StartServer()
}
