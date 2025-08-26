package readiness

type service struct {
	id string
	cb func() error
}

func (s *service) IsReady() error {
	return s.cb()
}

func (s *service) ID() string {
	return s.id
}

func NewService(id string, cb func() error) Service {
	return &service{
		id: id,
		cb: cb,
	}
}
