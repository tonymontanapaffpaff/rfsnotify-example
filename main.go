package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dietsche/rfsnotify"
	"github.com/fsnotify/fsnotify"
)

func waitUntilFind(target string) error {
	for {
		time.Sleep(1 * time.Second)
		_, err := os.Stat(target)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return err
			}
		}
		break
	}
	return nil
}

func watch(watcher *rfsnotify.RWatcher, removeCh chan bool, renameCh chan bool, errCh chan error) {
	for {
		select {
		case event := <-watcher.Events:
			switch {
			case fsnotify.Op(event.Op)&fsnotify.Create == fsnotify.Create:
				log.Println("created file:", event.Name)
			case fsnotify.Op(event.Op)&fsnotify.Write == fsnotify.Write:
				log.Println("modified file:", event.Name)
			case fsnotify.Op(event.Op)&fsnotify.Remove == fsnotify.Remove:
				log.Println("removed file:", event.Name)
				removeCh <- true
			case fsnotify.Op(event.Op)&fsnotify.Rename == fsnotify.Rename:
				log.Println("renamed file:", event.Name)
				renameCh <- true
			case fsnotify.Op(event.Op)&fsnotify.Chmod == fsnotify.Chmod:
				log.Println("mode changed file:", event.Name)
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
			errCh <- err
		}
	}
}

func synchronize(watcher *rfsnotify.RWatcher, target string, removeCh chan bool, renameCh chan bool) {
	for {
		select {
		case <-renameCh:
			err := waitUntilFind(target)
			if err != nil {
				log.Fatalln(err)
			}
			err = watcher.AddRecursive(target)
			if err != nil {
				log.Fatalln(err)
			}
		case <-removeCh:
			err := waitUntilFind(target)
			if err != nil {
				log.Fatalln(err)
			}
			err = watcher.AddRecursive(target)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func main() {
	// it watches the working directory if target not specified
	var target string
	flag.StringVar(&target, "t", "", "file or directory to be monitored")

	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	target = filepath.FromSlash(dir + "/" + target)

	err = waitUntilFind(target)
	if err != nil {
		log.Fatalln(err)
	}

	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.AddRecursive(target)
	if err != nil {
		log.Fatal(err)
	}

	removeCh := make(chan bool)
	renameCh := make(chan bool)
	errCh := make(chan error)

	go watch(watcher, removeCh, renameCh, errCh)
	go synchronize(watcher, target, removeCh, renameCh)

	log.Fatal(<-errCh)
}
