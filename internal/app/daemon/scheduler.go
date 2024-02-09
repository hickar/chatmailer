package daemon

import "time"

type Scheduler struct {
	callback func()
	ticker   *time.Ticker
	quit     chan struct{}
}

func (s *Scheduler) Schedule(callback func(), period time.Duration) {
	s.quit = make(chan struct{})
	s.callback = callback
	s.ticker = time.NewTicker(time.Minute)
	go s.runSchedule()
}

func (s *Scheduler) runSchedule() {
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

func (s *Scheduler) Stop() {
	close(s.quit)
}
