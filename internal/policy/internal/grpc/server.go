package grpc

import (
	"context"
	"net/http"
	"strconv"

	policyv1 "github.com/Duke1616/ecmdb/api/proto/gen/ecmdb/policy/v1"
	"github.com/Duke1616/ecmdb/internal/policy/internal/service"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PolicyServer struct {
	policyv1.UnimplementedPolicyServiceServer

	policySvc service.Service
	sp        session.Provider
}

func NewPolicyServer(policySvc service.Service, sp session.Provider) *PolicyServer {
	return &PolicyServer{
		policySvc: policySvc,
		sp:        sp,
	}
}

func (f *PolicyServer) Register(server grpc.ServiceRegistrar) {
	policyv1.RegisterPolicyServiceServer(server, f)
}

// CheckLogin 登录验证
func (f *PolicyServer) CheckLogin(ctx context.Context, req *policyv1.CheckLoginReq) (*policyv1.CheckLoginRes, error) {
	sess, err := f.getSession(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "登录认证失败: %v", err)
	}

	return &policyv1.CheckLoginRes{
		Uid: sess.Claims().Uid,
	}, nil
}

// Authorize 权限鉴权
func (f *PolicyServer) Authorize(ctx context.Context, req *policyv1.AuthorizeReq) (
	*policyv1.Response, error) {
	sess, err := f.getSession(req.Token)
	if err != nil {
		return &policyv1.Response{Allowed: false, Reason: "身份验证失败"}, nil
	}

	userId := strconv.FormatInt(sess.Claims().Uid, 10)
	result, err := f.policySvc.Authorize(ctx, userId, req.Path, req.Method, req.Resource)
	if err != nil {
		return &policyv1.Response{Allowed: false, Reason: err.Error()}, err
	}

	return &policyv1.Response{
		Allowed:         result.Allowed,
		Roles:           result.Roles,
		MatchedPolicies: result.MatchedPolicies,
		Reason:          result.Reason,
	}, nil
}

// getSession 通过 Token 获取 Session
// NOTE: session.Provider.Get 要求 *gctx.Context，
// TokenCarrier 固定从 Request.Header["Authorization"] 提取 Token，
// 所以这里构造一个仅包含 Authorization Header 的最小请求即可
func (f *PolicyServer) getSession(token string) (session.Session, error) {
	req, _ := http.NewRequest("", "/", nil)
	req.Header.Set("Authorization", token)
	return f.sp.Get(&gctx.Context{Context: &gin.Context{Request: req}})
}
