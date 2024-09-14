package web

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/policy"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/Duke1616/ecmdb/internal/user/ldapx"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc           service.Service
	ldapSvc       service.LdapService
	policySvc     policy.Service
	departmentSvc department.Service
}

func NewHandler(svc service.Service, ldapSvc service.LdapService,
	policySvc policy.Service, departmentSvc department.Service) *Handler {
	return &Handler{
		svc:           svc,
		ldapSvc:       ldapSvc,
		policySvc:     policySvc,
		departmentSvc: departmentSvc,
	}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/user")
	g.POST("/ldap/login", ginx.WrapBody[LoginLdapReq](h.LoginLdap))
	g.POST("/system/login", ginx.WrapBody[LoginSystemReq](h.LoginSystem))
	g.POST("/refresh", ginx.Wrap(h.RefreshAccessToken))
	g.POST("/register", ginx.WrapBody[RegisterUserReq](h.RegisterUser))
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/user")
	g.POST("/role/bind", ginx.WrapBody[UserBindRoleReq](h.UserRoleBind))
	g.POST("/list", ginx.WrapBody[Page](h.ListUser))
	g.POST("/update", ginx.WrapBody[UpdateUserReq](h.UpdateUser))
	g.POST("/info", ginx.Wrap(h.GetUserInfo))
	g.POST("/find/usernames", ginx.WrapBody[FindByUserNamesReq](h.FindByUsernames))
	g.POST("/pipeline/department_id", ginx.Wrap(h.PipelineDepartmentId))
	g.POST("/find/regex/username", ginx.WrapBody[FindByUsernameRegexReq](h.FindByUsernameRegex))
	g.POST("/find/department_id", ginx.WrapBody[FindUsersByDepartmentIdReq](h.FindByDepartmentId))
}

func (h *Handler) LoginSystem(ctx *gin.Context, req LoginSystemReq) (ginx.Result, error) {
	user, err := h.svc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return userOrPassErrorResult, err
	}

	jwtData := make(map[string]string, 0)
	_, err = session.NewSessionBuilder(&gctx.Context{Context: ctx}, user.Id).SetJwtData(jwtData).Build()
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.ToUserVo(user),
		Msg:  "登录用户成功",
	}, nil
}

func (h *Handler) FindByUsernames(ctx *gin.Context, req FindByUserNamesReq) (ginx.Result, error) {
	if len(req.Usernames) < 0 {
		return systemErrorResult, fmt.Errorf("输入为空，不符合要求")
	}

	users, err := h.svc.FindByUsernames(ctx, req.Usernames)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询指定多个用户的详情信息",
		Data: RetrieveUsers{
			Total: int64(len(users)),
			Users: slice.Map(users, func(idx int, src domain.User) User {
				return h.ToUserVo(src)
			}),
		},
	}, nil
}

func (h *Handler) PipelineDepartmentId(ctx *gin.Context) (ginx.Result, error) {
	// 根据 组ID 聚合查询所有数据
	pipeline, err := h.svc.PipelineDepartmentId(ctx)
	if err != nil {
		return systemErrorResult, err
	}

	deps, err := h.departmentSvc.ListDepartment(ctx)
	if err != nil {
		return systemErrorResult, err
	}

	data, err := GenerateDepartmentUserTree(pipeline, deps)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: data,
	}, nil
}

func (h *Handler) UpdateUser(ctx *gin.Context, req UpdateUserReq) (ginx.Result, error) {
	id, err := h.svc.UpdateUser(ctx, h.toUpdateDomain(req))
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: id,
	}, nil
}

func (h *Handler) FindByDepartmentId(ctx *gin.Context, req FindUsersByDepartmentIdReq) (ginx.Result, error) {
	users, total, err := h.svc.FindByDepartmentId(ctx, req.Offset, req.Limit, req.DepartmentId)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询部门组内用户列表成功",
		Data: RetrieveUsers{
			Total: total,
			Users: slice.Map(users, func(idx int, src domain.User) User {
				return h.ToUserVo(src)
			}),
		},
	}, nil
}

func (h *Handler) FindByUsernameRegex(ctx *gin.Context, req FindByUsernameRegexReq) (ginx.Result, error) {
	users, total, err := h.svc.FindByUsernameRegex(ctx, req.Offset, req.Limit, req.Username)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询指定用户列表成功",
		Data: RetrieveUsers{
			Total: total,
			Users: slice.Map(users, func(idx int, src domain.User) User {
				return h.ToUserVo(src)
			}),
		},
	}, nil
}

func (h *Handler) RegisterUser(ctx *gin.Context, req RegisterUserReq) (ginx.Result, error) {
	// 两次密码输入不一致
	if req.Password != req.RePassword {
		return systemErrorResult, nil
	}

	// 查询用户并创建
	u, err := h.svc.FindOrCreateBySystem(ctx, req.Username, req.Password, req.DisplayName)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.ToUserVo(u),
	}, nil
}

func (h *Handler) LoginLdap(ctx *gin.Context, req LoginLdapReq) (ginx.Result, error) {
	// LDAP 登录成功
	profile, err := h.ldapSvc.Login(ctx, domain.User{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return userOrPassErrorResult, err
	}

	// 查找或插入用户
	user, err := h.svc.FindOrCreateByLdap(ctx, h.toDomain(profile))
	if err != nil {
		return systemErrorResult, err
	}

	// 生成session
	jwtData := make(map[string]string, 0)
	_, err = session.NewSessionBuilder(&gctx.Context{Context: ctx}, user.Id).SetJwtData(jwtData).Build()
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.ToUserVo(user),
		Msg:  "登录用户成功",
	}, nil
}

func (h *Handler) ListUser(ctx *gin.Context, req Page) (ginx.Result, error) {
	rts, total, err := h.svc.ListUser(ctx, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "查询用户列表成功",
		Data: RetrieveUsers{
			Total: total,
			Users: slice.Map(rts, func(idx int, src domain.User) User {
				return h.ToUserVo(src)
			}),
		},
	}, nil
}

func (h *Handler) UserRoleBind(ctx *gin.Context, req UserBindRoleReq) (ginx.Result, error) {
	_, err := h.svc.AddRoleBind(ctx, req.Id, req.RoleCodes)
	if err != nil {
		return systemErrorResult, err
	}

	ok, err := h.policySvc.UpdateFilteredGrouping(ctx, req.Id, req.RoleCodes)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: ok,
		Msg:  "用户角色绑定成功",
	}, nil
}

func (h *Handler) RefreshAccessToken(ctx *gin.Context) (ginx.Result, error) {
	err := session.RenewAccessToken(&gctx.Context{Context: ctx})
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (h *Handler) GetUserInfo(ctx *gin.Context) (ginx.Result, error) {
	// 获取登录用户 sess 获取ID
	sess, err := session.Get(&gctx.Context{Context: ctx})
	if err != nil {
		return systemErrorResult, fmt.Errorf("获取 Session 失败, %w", err)
	}

	user, err := h.svc.FindById(ctx, sess.Claims().Uid)
	if err != nil {
		return ginx.Result{}, err
	}

	return ginx.Result{
		Data: h.ToUserVo(user),
	}, nil
}

func (h *Handler) toDomain(profile *ldapx.Profile) domain.User {
	return domain.User{
		Username:    profile.Username,
		Email:       profile.Email,
		Title:       profile.Title,
		DisplayName: profile.DisplayName,
		Status:      domain.ENABLED,
		CreateType:  domain.LDAP,
	}
}

func (h *Handler) toUpdateDomain(req UpdateUserReq) domain.User {
	return domain.User{
		Id:           req.Id,
		Email:        req.Email,
		Title:        req.Title,
		DisplayName:  req.DisplayName,
		DepartmentId: req.DepartmentId,
	}
}

func (h *Handler) ToUserVo(src domain.User) User {
	return User{
		Id:           src.Id,
		DepartmentId: src.DepartmentId,
		Username:     src.Username,
		Email:        src.Email,
		Title:        src.Title,
		RoleCodes:    src.RoleCodes,
		DisplayName:  src.DisplayName,
		CreateType:   src.CreateType.ToUint8(),
	}
}
