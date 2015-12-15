package material

import (
	"log"
	"path/filepath"
	"strings"

	"gopkg.in/fsnotify.v1"
)

func watchShaders() (chan string, chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	event := make(chan string)
	quit := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				log.Println(ev)
				// expects filenames such as triangle-vert.glsl or triangle-frag.glsl
				if filepath.Ext(ev.Name) == ".glsl" && ev.Op == fsnotify.Write {
					b := filepath.Base(ev.Name)
					i := strings.LastIndex(b, "-")
					b = b[:i]
					event <- b
				}
			case err := <-watcher.Errors:
				log.Println("watch error:", err)
			case <-quit:
				watcher.Close()
				return
			}
		}
	}()

	err = watcher.Add("./assets")
	if err != nil {
		log.Fatal("Failed to watch folder.", err)
	}

	return event, quit
}
