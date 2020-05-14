package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
	"github.com/namsral/flag"
)

var (
	crontabPath string
	numCPU      int
)

func init() {
	flag.StringVar(&crontabPath, "file", "crontab", "crontab file path")
	flag.IntVar(&numCPU, "cpu", runtime.NumCPU(), "maximum number of CPUs")
}

func reload() (*Runner, error) {
	file, err := os.Open(crontabPath)
	if err != nil {
		return nil, fmt.Errorf("crontab path:%v err:%v", crontabPath, err)
	}

	parser, err := NewParser(file)
	if err != nil {
		return nil, fmt.Errorf("Parser read err:%v", err)
	}

	runner, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("Parser parse err:%v", err)
	}

	file.Close()

	return runner, nil
}

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(numCPU)

	runner, err := reload()
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Watcher err:%v", err)
	}

	debounced := debounce.New(10 * time.Second)

	var wg sync.WaitGroup

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			select {

			case evt, ok := <-watcher.Events:
				if !ok {
					continue
				}
				// we're watching one directory for one file; this simplification solves
				// all kinds of weird filesystem/os complexity.
				if filepath.Base(evt.Name) != filepath.Base(crontabPath) {
					continue
				}
				log.Println("Crontab changed, reloading...")
				debounced(func() {
					newRunner, err := reload()
					if err != nil {
						log.Printf("Error on reload, ignoring (no changes were applied): %v", err)
						return
					}
					newRunner.Start()
					runner.Stop()
					runner = newRunner
				})

			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				log.Fatalf("Watcher err:%v", err)

			case sig := <-c:
				log.Println("Got signal: ", sig)
				watcher.Close()
				runner.Stop()
				wg.Done()
			}
		}
	}()

	runner.Start()
	wg.Add(1)

	if err := watcher.Add(filepath.Dir(crontabPath)); err != nil {
		log.Fatalf("Watcher err:%v", err)
	}

	wg.Wait()
	log.Println("End cron")
}
