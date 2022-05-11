package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/dietsche/rfsnotify"
	"github.com/fsnotify/fsnotify"
)

func watch(watcher *rfsnotify.RWatcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			switch {
			case fsnotify.Op(event.Op)&fsnotify.Create == fsnotify.Create:
				log.Println("created file:", event.Name)
			case fsnotify.Op(event.Op)&fsnotify.Write == fsnotify.Write:
				log.Println("modified file:", event.Name)
			case fsnotify.Op(event.Op)&fsnotify.Remove == fsnotify.Remove:
				log.Println("removed file:", event.Name)
			case fsnotify.Op(event.Op)&fsnotify.Rename == fsnotify.Rename:
				log.Println("renamed file:", event.Name)
			case fsnotify.Op(event.Op)&fsnotify.Chmod == fsnotify.Chmod:
				log.Println("mode changed file:", event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func main() {
	var target string
	flag.StringVar(&target, "t", "", "file or directory to be monitored")

	flag.Parse()

	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go watch(watcher)

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.FromSlash(dir + "/" + target)
	err = watcher.AddRecursive(path)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
