package frontend_flow

type FrontendFlow interface {
	Deploy() (int, error)
}
