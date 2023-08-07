package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/radovskyb/watcher"
)

const RELOAD_COMMAND = "r"

func runFlutter() *exec.Cmd {
	args := append([]string{"run"}, os.Args[1:]...)
	cmd := exec.Command("flutter", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func runWatcher(writer io.WriteCloser) {
	log.Println("starting watcher...")
	payload := []byte(RELOAD_COMMAND)

	w := watcher.New()
	w.SetMaxEvents(1)
	r := regexp.MustCompile("^*.dart")
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		for {
			select {
			case _ = <-w.Event:
				log.Println("file change event: sending reload request.")
				writer.Write(payload)
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive("."); err != nil {
		log.Fatalln(err)
	}

	if err := w.Start(time.Millisecond * 500); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	// started run.
	flutterCmd := runFlutter()
	flutterInPipe, _ := flutterCmd.StdinPipe()

	go runWatcher(flutterInPipe)

	err := flutterCmd.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
