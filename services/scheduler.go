package services

import (
	"time"

	"github.com/go-co-op/gocron"
)

var Scheduler *gocron.Scheduler

func GetScheduler() *gocron.Scheduler {
	if Scheduler == nil {
		Scheduler = gocron.NewScheduler(time.UTC).SingletonMode() // a long running job will not be rescheduled until the current run is completed
	}
	return Scheduler
}

func StartInBackground() {
	s := GetScheduler()
	s.StartAsync()
}
