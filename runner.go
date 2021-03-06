package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"os/exec"
)

type Runner struct {
	cron *cron.Cron
}

func NewRunner() *Runner {
	r := &Runner{
		cron: cron.New(),
	}
	return r
}

func (r *Runner) Add(spec string, cmd string) error {
	entryID, err := r.cron.AddFunc(spec, r.cmdFunc(cmd))

	log.Printf("Add cron job spec:%v cmd:%v entryID: %v err:%v", spec, cmd, entryID, err)

	return err
}

func (r *Runner) Len() int {
	return len(r.cron.Entries())
}

func (r *Runner) Start() {
	log.Println("Start runner")
	r.cron.Start()
}

func (r *Runner) Stop() {
	r.cron.Stop()
	log.Println("Stop runner")
}

func (r *Runner) cmdFunc(cmd string) func() {
	cmdFunc := func() {
		out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
		log.Printf("cmd:%v out:%v err:%v", cmd, string(out), err)
	}
	return cmdFunc
}
