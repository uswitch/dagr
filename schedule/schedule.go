package schedule

import (
	"github.com/robfig/cron"
	"log"
)

type Schedule struct {
	cron.Schedule
}

var DefaultSchedule Schedule

func init() {
	cronSchedule, err := cron.Parse("@daily")
	if err != nil {
		panic(err)
	}

	DefaultSchedule = Schedule{cronSchedule}
}

func (s *Schedule) UnmarshalText(text []byte) error {
	var err error
	cronSchedule, err := cron.Parse(string(text))
	log.Println("parse schedule: ", cronSchedule, string(text), err)
	s.Schedule = cronSchedule
	return err
}
