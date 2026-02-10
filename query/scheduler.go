package query

type Scheduler interface {
	ScheduleAfter(delay uint64, cb func())
}

type MockScheduler struct {
	callbacks []func()
}

func (s *MockScheduler) ScheduleAfter(delay uint64, cb func()) {
	s.callbacks = append(s.callbacks, cb)
}

func (s *MockScheduler) FireAll() {
	for _, cb := range s.callbacks {
		cb()
	}
	s.callbacks = nil
}
