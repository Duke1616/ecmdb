package endpoint

type Module struct {
	Hdl       *Handler
	Svc       Service
	RpcServer *RpcServer
}
