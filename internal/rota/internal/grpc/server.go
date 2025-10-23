package grpc

import (
	"context"

	rotav1 "github.com/Duke1616/ecmdb/api/proto/gen/rota/v1"
	"github.com/Duke1616/ecmdb/internal/rota/internal/domain"
	"github.com/Duke1616/ecmdb/internal/rota/internal/service"
	"github.com/ecodeclub/ekit/slice"

	"google.golang.org/grpc"
)

type RotaServer struct {
	rotav1.UnimplementedOnCallServiceServer

	rotaSvc service.Service
}

func (u *RotaServer) GetCurrentSchedule(ctx context.Context, req *rotav1.GetCurrentScheduleRequest) (*rotav1.Schedule, error) {
	scheduler, err := u.rotaSvc.GetCurrentSchedule(ctx, req.Id)
	return u.ToRetrieveUsers(scheduler), err
}

func (u *RotaServer) GetCurrentSchedulesByIDs(ctx context.Context, req *rotav1.GetCurrentSchedulesByIDsRequest) (*rotav1.Schedules, error) {
	schedules, err := u.rotaSvc.GetCurrentSchedulesByIDs(ctx, req.Ids)
	if err != nil {
		return nil, err
	}

	return &rotav1.Schedules{
		Schedules: slice.Map(schedules, func(idx int, src domain.Schedule) *rotav1.Schedule {
			return u.ToRetrieveUsers(src)
		}),
	}, nil
}

func NewRotaServer(rotaSvc service.Service) *RotaServer {
	return &RotaServer{rotaSvc: rotaSvc}
}

func (u *RotaServer) Register(server grpc.ServiceRegistrar) {
	rotav1.RegisterOnCallServiceServer(server, u)
}

func (u *RotaServer) ToRetrieveUsers(src domain.Schedule) *rotav1.Schedule {
	return &rotav1.Schedule{
		Title: src.Title,
		RotaGroup: &rotav1.RotaGroup{
			Id:      src.RotaGroup.Id,
			Name:    src.RotaGroup.Name,
			Members: src.RotaGroup.Members,
		},
		StartTime: src.StartTime,
		EndTime:   src.EndTime,
	}
}
