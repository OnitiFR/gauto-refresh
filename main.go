package main

import (
	"flag"
	"fmt"
)

type FlagValues []string

const Version = "0.0.3"

var ListenPort string
var Files FlagValues

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

	fmt.Println("Watching", Files)
	fmt.Println("Add the follow line to your HTML page:")
	fmt.Println("--")
	fmt.Printf("<script src=\"http://%s/refresh\"></script>\n", ListenPort)
	fmt.Println("--")
	StartServer()
}
