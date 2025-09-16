package resource

type Module struct {
	Svc          Service
	EncryptedSvc EncryptedSvc
	Hdl          *Handler
}
