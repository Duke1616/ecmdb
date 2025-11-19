package model

type Module struct {
	Svc   Service
	MGSvc MGService
	Hdl   *Handler
}
