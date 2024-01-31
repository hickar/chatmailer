package main

import "time"

type scheduler struct {
	callback func()
	ticker   *time.Ticker
	quit     chan struct{}
}

func (s *scheduler) Schedule(callback func(), period time.Duration) {
	s.quit = make(chan struct{})
	s.callback = callback
	s.ticker = time.NewTicker(time.Minute)
	go s.runSchedule()
}

func (s *scheduler) runSchedule() {
	defer s.ticker.Stop()

	for {
		select {
		case <-s.ticker.C:
			s.callback()
		case <-s.quit:
			return
		}
	}
}

func (s *scheduler) Stop() {
	close(s.quit)
}
