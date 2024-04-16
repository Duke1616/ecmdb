package relation

type Module struct {
	RRSvc RRSvc
	RMSvc RMSvc
	RTSvc RTSvc
	RRHdl *RRHandler
	RMHdl *RMHandler
	RTHdl *RTHandler
}
