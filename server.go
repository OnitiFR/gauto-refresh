package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var broker *Broker[bool]

var exclude = []string{
	".git",
	".svn",
	"node_modules",
	"vendor",
}

func serveScript(w http.ResponseWriter, req *http.Request) {
	str := `
	const es = new EventSource(sse_url)
	es.addEventListener("ask-refresh", (ev) => {
		###
	})
	`
	str = strings.Replace(str, "###", Action, 1)

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprintf(w, "var sse_url='http://%s/sse'", ListenPort)
	w.Write([]byte(str))
}

func serveSSE(w http.ResponseWriter, req *http.Request) {

	if req.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		return
	}

	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("client connected")

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	msgCh := broker.Subscribe()
	defer broker.Unsubscribe(msgCh)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	refreshRequest := make(chan bool, 1)
	refreshRequest <- true

	done := make(chan bool)
	go func() {
		for {
			select {
			case <-msgCh:
				go func() {
					select {
					case <-refreshRequest:
					default:
						DebugLog("already asked for refresh, ignore")
						return
					}

					time.Sleep(time.Duration(Delay) * time.Millisecond)
					log.Println("ask client refresh")
					fmt.Fprintf(w, "event: ask-refresh\n")
					fmt.Fprintf(w, "data: {}\n")
					fmt.Fprintf(w, "\n\n")
					flusher.Flush()

					refreshRequest <- true
				}()
			case <-req.Context().Done():
				done <- true
			}
		}
	}()
	<-done
	log.Println("client disconnected")
}

func startWatchers() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for _, file := range Files {
		DebugLog("from '%s' …", file)
		stat, err := os.Stat(file)
		if err != nil {
			log.Fatal(err)
		}

		if stat.Mode().IsDir() {
			filepath.Walk(file, func(path string, fi fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if fi.Mode().IsDir() {
					for _, e := range exclude {
						if fi.Name() == e {
							DebugLog("… skip dir '%s'", path)
							return filepath.SkipDir
						}
					}

					err := watcher.Add(path)
					DebugLog("… watching dir '%s'", path)
					if err != nil {
						log.Printf("error: %s", err)
					}
				}
				return nil
			})
		} else {
			err = watcher.Add(file)
			DebugLog("… watching file '%s'", file)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Println("event:", event)

			if event.Op&fsnotify.Write == fsnotify.Write {
				DebugLog("change detected")
				broker.Publish(true)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}

}

func StartServer() {
	broker = NewBroker[bool]()
	go broker.Start()

	go startWatchers()

	http.HandleFunc("/sse", serveSSE)
	http.HandleFunc("/refresh", serveScript)
	DebugLog("listening on %s", ListenPort)
	err := http.ListenAndServe(ListenPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
