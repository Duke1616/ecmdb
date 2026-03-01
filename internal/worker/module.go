package worker

import (
	"github.com/Duke1616/ecmdb/internal/worker/internal/job"
)

type Module struct {
	Svc Service
	Job *job.ServiceDiscoveryJob
}
