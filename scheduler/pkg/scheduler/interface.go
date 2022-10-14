package scheduler

type Scheduler interface {
	Schedule(modelKey string) error
	ScheduleFailedModels() ([]string, error)
}
