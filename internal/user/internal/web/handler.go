package web

import (
	"fmt"

	"github.com/Duke1616/ecmdb/internal/department"
	"github.com/Duke1616/ecmdb/internal/user/internal/domain"
	"github.com/Duke1616/ecmdb/internal/user/internal/service"
	"github.com/Duke1616/ecmdb/pkg/ginx"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc           service.Service
	ldapSvc       service.LdapService
	sp            session.Provider
	departmentSvc department.Service
}

func NewHandler(svc service.Service, ldapSvc service.LdapService,
	departmentSvc department.Service, sp session.Provider) *Handler {
	return &Handler{
		svc:           svc,
		ldapSvc:       ldapSvc,
		departmentSvc: departmentSvc,
		sp:            sp,
	}
}

func (h *Handler) PublicRoutes(server *gin.Engine) {
	g := server.Group("/api/user")
	g.POST("/ldap/login", ginx.WrapBody[LoginLdapReq](h.LoginLdap))
	g.POST("/system/login", ginx.WrapBody[LoginSystemReq](h.LoginSystem))
	g.POST("/refresh", ginx.Wrap(h.RefreshAccessToken))
	g.POST("/register", ginx.WrapBody[RegisterUserReq](h.RegisterUser))
	g.POST("/logout", ginx.Wrap(h.Logout))
}

func (h *Handler) PrivateRoutes(server *gin.Engine) {
	g := server.Group("/api/user")
	g.POST("/role/bind", ginx.WrapBody[UserBindRoleReq](h.UserRoleBind))
	g.POST("/list", ginx.WrapBody[Page](h.ListUser))
	g.POST("/update", ginx.WrapBody[UpdateUserReq](h.UpdateUser))
	g.POST("/info", ginx.Wrap(h.GetUserInfo))
	g.POST("/find/usernames", ginx.WrapBody[FindByUserNamesReq](h.FindByUsernames))
	g.POST("/pipeline/department_id", ginx.Wrap(h.PipelineDepartmentId))
	g.POST("/find/by_keyword", ginx.WrapBody[FindByKeywordReq](h.FindByKeywords))
	g.POST("/find/username", ginx.WrapBody[FindByUserNameReq](h.FindByUsername))
	g.POST("/find/id", ginx.WrapBody[FindByIdReq](h.FindById))
	g.POST("/find/by_ids", ginx.WrapBody[FindByIdsReq](h.FindByIds))
	g.POST("/find/department_id", ginx.WrapBody[FindUsersByDepartmentIdReq](h.FindByDepartmentId))

	// 查询 LDAP 用户
	g.POST("/ldap/search", ginx.WrapBody[SearchLdapUser](h.SearchLdapUser))
	g.POST("/ldap/sync", ginx.WrapBody[SyncLdapUserReq](h.SyncLdapUser))
	g.POST("/ldap/refresh_cache", ginx.Wrap(h.LdapRefreshCache))
}

func (h *Handler) Logout(ctx *gin.Context) (ginx.Result, error) {
	err := h.sp.Destroy(&gctx.Context{
		Context: ctx,
	})

	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *Handler) SyncLdapUser(ctx *gin.Context, req SyncLdapUserReq) (ginx.Result, error) {
	userId, err := h.svc.SyncCreateLdapUser(ctx, domain.User{
		DepartmentId: req.DepartmentId,
		Username:     req.Username,
		Email:        req.Email,
		Title:        req.Title,
		DisplayName:  req.DisplayName,
		Status:       domain.ENABLED,
		CreateType:   domain.LDAP,
		RoleCodes:    req.RoleCodes,
		FeishuInfo: domain.FeishuInfo{
			UserId: req.FeishuInfo.UserId,
		},
		WechatInfo: domain.WechatInfo{
			UserId: req.WechatInfo.UserId,
		},
	})

	if err != nil {
		return systemErrorResult, nil
	}

	return ginx.Result{
		Data: userId,
		Msg:  "录入用户成功",
	}, nil
}

func (h *Handler) FindByIds(ctx *gin.Context, req FindByIdsReq) (ginx.Result, error) {
	if len(req.Ids) < 0 {
		return systemErrorResult, fmt.Errorf("输入为空，不符合要求")
	}

	users, err := h.svc.FindByIds(ctx, req.Ids)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Msg: "获取多个用户信息成功",
		Data: RetrieveUsers{
			Total: int64(len(users)),
			Users: slice.Map(users, func(idx int, src domain.User) User {
				return h.ToUserVo(src)
			}),
		},
	}, nil
}

func (h *Handler) SearchLdapUser(ctx *gin.Context, req SearchLdapUser) (ginx.Result, error) {
	// 这个是全量的数据查询
	pager, total, err := h.ldapSvc.SearchUserWithPager(ctx, req.Keywords, req.Offset, req.Limit)
	if err != nil {
		return systemErrorResult, err
	}

	// 查询本地数据库，去除已经插入的用户
	us := slice.Map(pager, func(idx int, src domain.Profile) string {
		return src.Username
	})

	users, err := h.svc.FindByUsernames(ctx, us)
	if err != nil {
		return systemErrorResult, err
	}

	uniqueMap := slice.ToMapV(users, func(element domain.User) (string, bool) {
		return element.Username, true
	})

	result := slice.Map(pager, func(idx int, src domain.Profile) LdapUser {
		isExist := false
		if _, ok := uniqueMap[src.Username]; ok {
			isExist = true
		}

		return LdapUser{
			Username:      src.Username,
			Email:         src.Email,
			Title:         src.Title,
			DisplayName:   src.DisplayName,
			IsSystemExist: isExist,
		}
	})

	return ginx.Result{
		Data: RetrieveLdapUsers{
			Total: total,
			Users: result,
		},
		Msg: "查询 LDAP 用户",
	}, nil
}

func (h *Handler) LdapRefreshCache(ctx *gin.Context) (ginx.Result, error) {
	err := h.ldapSvc.RefreshCacheUserWithPager(ctx)
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: "ok",
		Msg:  "刷新缓存成功",
	}, nil
}

func (h *Handler) LoginSystem(ctx *gin.Context, req LoginSystemReq) (ginx.Result, error) {
	user, err := h.svc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return userOrPassErrorResult, err
	}

	jwtData := map[string]string{
		"username": user.Username,
	}
	_, err = session.NewSessionBuilder(&gctx.Context{Context: ctx}, user.Id).SetJwtData(jwtData).Build()
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.ToUserVo(user),
		Msg:  "登录用户成功",
	}, nil
}

func (h *Handler) FindByUsername(ctx *gin.Context, req FindByUserNameReq) (ginx.Result, error) {
	var u User
	if req.Username == "" {
		sess, err := h.sp.Get(&gctx.Context{Context: ctx})
		if err != nil {
			return systemErrorResult, fmt.Errorf("获取 Session 失败, %w", err)
		}

		user, err := h.svc.FindById(ctx, sess.Claims().Uid)
		if err != nil {
			return systemErrorResult, err
		}

		u = h.ToUserVo(user)
	} else {
		user, err := h.svc.FindByUsername(ctx, req.Username)
		if err != nil {
			return systemErrorResult, err
		}

		u = h.ToUserVo(user)
	}

	return ginx.Result{
		Data: u,
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

func (h *Handler) FindById(ctx *gin.Context, req FindByIdReq) (ginx.Result, error) {
	var u User
	if req.Id == 0 {
		sess, err := h.sp.Get(&gctx.Context{Context: ctx})
		if err != nil {
			return systemErrorResult, fmt.Errorf("获取 Session 失败, %w", err)
		}

		user, err := h.svc.FindById(ctx, sess.Claims().Uid)
		if err != nil {
			return systemErrorResult, err
		}

		u = h.ToUserVo(user)
	} else {
		user, err := h.svc.FindById(ctx, req.Id)
		if err != nil {
			return systemErrorResult, err
		}

		u = h.ToUserVo(user)
	}

	return ginx.Result{
		Data: u,
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

func (h *Handler) FindByKeywords(ctx *gin.Context, req FindByKeywordReq) (ginx.Result, error) {
	users, total, err := h.svc.FindByKeywords(ctx, req.Offset, req.Limit, req.Keyword)
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
	u, err := h.svc.FindOrCreateBySystem(ctx, h.ToRegisterVo(req))
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
	jwtData := map[string]string{
		"username": user.Username,
	}
	_, err = session.NewSessionBuilder(&gctx.Context{Context: ctx}, user.Id).SetJwtData(jwtData).Build()
	if err != nil {
		return systemErrorResult, err
	}

	return ginx.Result{
		Data: h.ToUserVo(user),
		Msg:  "登录用户成功",
	}, nil
}

func (h *Handler) RefreshAccessToken(ctx *gin.Context) (ginx.Result, error) {
	err := h.sp.RenewAccessToken(&gctx.Context{Context: ctx})
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (h *Handler) ListUser(ctx *gin.Context, req Page) (ginx.Result, error) {
	// 设置分页默认值
	offset := req.Offset
	limit := req.Limit
	if limit <= 0 {
		limit = 10 // 默认每页10条
	}
	if offset < 0 {
		offset = 0
	}

	rts, total, err := h.svc.ListUser(ctx, offset, limit)
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

	return ginx.Result{
		Data: "ok",
		Msg:  "用户角色绑定成功",
	}, nil
}

func (h *Handler) GetUserInfo(ctx *gin.Context) (ginx.Result, error) {
	// 获取登录用户 sess 获取ID
	sess, err := h.sp.Get(&gctx.Context{Context: ctx})
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

func (h *Handler) toDomain(profile domain.Profile) domain.User {
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
		FeishuInfo: domain.FeishuInfo{
			UserId: req.FeishuInfo.UserId,
		},
		WechatInfo: domain.WechatInfo{
			UserId: req.WechatInfo.UserId,
		},
	}
}

func (h *Handler) ToRegisterVo(req RegisterUserReq) domain.User {
	return domain.User{
		Email:        req.Email,
		Title:        req.Title,
		Password:     req.Password,
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		DepartmentId: req.DepartmentId,
		FeishuInfo: domain.FeishuInfo{
			UserId: req.FeishuInfo.UserId,
		},
		WechatInfo: domain.WechatInfo{
			UserId: req.WechatInfo.UserId,
		},
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
		FeishuInfo: FeishuInfo{
			UserId: src.FeishuInfo.UserId,
		},
		WechatInfo: WechatInfo{
			UserId: src.WechatInfo.UserId,
		},
	}
}
