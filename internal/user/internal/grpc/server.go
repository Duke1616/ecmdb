package grpc

import (
	"context"

	userv1 "github.com/Duke1616/ecmdb/api/proto/gen/ecmdb/user/v1"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer

	userSvc service.Service
}

func (u *UserServer) FindByUsernames(ctx context.Context, req *userv1.FindByUsernamesReq) (*userv1.RetrieveUsers, error) {
	userInfos, err := u.userSvc.FindByUsernames(ctx, req.Usernames)
	return &userv1.RetrieveUsers{Users: slice.Map(userInfos, func(idx int, src domain.User) *userv1.User {
		return u.ToRetrieveUsers(src)
	})}, err
}

func (u *UserServer) FindByIds(ctx context.Context, req *userv1.FindByIdsReq) (*userv1.RetrieveUsers, error) {
	userInfos, err := u.userSvc.FindByIds(ctx, req.Ids)
	return &userv1.RetrieveUsers{Users: slice.Map(userInfos, func(idx int, src domain.User) *userv1.User {
		return u.ToRetrieveUsers(src)
	})}, err
}

func (u *UserServer) FindByDepartmentId(ctx context.Context, req *userv1.FindByDepartmentIdReq) (*userv1.RetrieveUsers, error) {
	userInfos, _, err := u.userSvc.FindByDepartmentId(ctx, 0, 20, req.DepartmentId)
	return &userv1.RetrieveUsers{Users: slice.Map(userInfos, func(idx int, src domain.User) *userv1.User {
		return u.ToRetrieveUsers(src)
	})}, err
}

func (u *UserServer) FindByDepartmentIds(ctx context.Context, req *userv1.FindByDepartmentIdsReq) (*userv1.RetrieveUsers, error) {
	userInfos, err := u.userSvc.FindByDepartmentIds(ctx, req.DepartmentIds)
	return &userv1.RetrieveUsers{Users: slice.Map(userInfos, func(idx int, src domain.User) *userv1.User {
		return u.ToRetrieveUsers(src)
	})}, err
}

func NewUserServer(userSvc service.Service) *UserServer {
	return &UserServer{userSvc: userSvc}
}

func (u *UserServer) Register(server grpc.ServiceRegistrar) {
	userv1.RegisterUserServiceServer(server, u)
}

func (u *UserServer) ToRetrieveUsers(src domain.User) *userv1.User {
	return &userv1.User{
		Id:           src.Id,
		Username:     src.Username,
		Email:        src.Email,
		LarkUserId:   src.FeishuInfo.UserId,
		WechatUserId: src.WechatInfo.UserId,
	}
}
