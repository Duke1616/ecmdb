package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/viper"
)

// SDK 提供登录验证和权限鉴权的 Gin 中间件
type SDK struct {
	baseURL string
	client  *http.Client
	logger  *elog.Component
}

// NewSDK 创建鉴权 SDK 实例，通过读取配置 policy.auth_url 获取 ECMDB 地址
// NOTE: 需要在配置文件中声明 policy.auth_url: http://ecmdb:8000
func NewSDK() *SDK {
	baseURL := viper.GetString("policy.auth_url")
	if baseURL == "" {
		panic("policy.auth_url 未配置，请在配置文件中声明 policy.auth_url")
	}
	return NewSDKWithURL(baseURL)
}

// NewSDKWithURL 创建鉴权 SDK 实例，显式传入 ECMDB 的地址
func NewSDKWithURL(baseURL string) *SDK {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8000" // 默认值
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &SDK{
		baseURL: baseURL,
		client:  &http.Client{},
		logger:  elog.DefaultLogger.With(elog.FieldComponentName("policy-sdk")),
	}
}

type checkLoginResp struct {
	Uid int64 `json:"uid"`
}

type checkPolicyReq struct {
	Path     string `json:"path"`
	Method   string `json:"method"`
	Resource string `json:"resource"`
}

type authorizeResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

type apiResult[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

// CheckLogin 登录检查中间件
func (s *SDK) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var res apiResult[checkLoginResp]
		if err := s.callAPI(ctx, "/api/policy/check_login", nil, &res); err != nil {
			return
		}

		ctx.Next()
	}
}

// CheckPolicy 权限鉴权中间件
func (s *SDK) CheckPolicy(resource string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var res apiResult[authorizeResult]
		if err := s.callAPI(ctx, "/api/policy/check_policy", checkPolicyReq{
			Path:     ctx.Request.URL.Path,
			Method:   ctx.Request.Method,
			Resource: resource,
		}, &res); err != nil {
			return
		}

		if !res.Data.Allowed {
			s.logger.Warn("用户无权限",
				elog.String("path", ctx.Request.URL.Path),
				elog.String("resource", resource),
				elog.String("reason", res.Data.Reason))
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		ctx.Next()
	}
}

// callAPI 向鉴权中心发送请求并解析响应，内部处理 Header 透传、错误处理、JSON 解码
func (s *SDK) callAPI(ctx *gin.Context, path string, body any, out any) error {
	var reqBody *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx.Request.Context(), "POST", s.baseURL+path, reqBody)
	if err != nil {
		s.logger.Error("创建请求失败", elog.FieldErr(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return err
	}

	// 直接复制原始请求的全部 Header，Authorization / Cookie 等自然透传
	req.Header = ctx.Request.Header.Clone()
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("请求鉴权中心失败", elog.FieldErr(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return err
	}
	defer resp.Body.Close()

	// 透传 Response Header（续期的新 Token 会通过 X-Access-Token / Set-Cookie 返回）
	for k, vals := range resp.Header {
		if k == "Content-Length" || k == "Content-Type" {
			continue
		}
		for _, v := range vals {
			ctx.Writer.Header().Add(k, v)
		}
	}

	if resp.StatusCode != http.StatusOK {
		ctx.AbortWithStatus(resp.StatusCode)
		return fmt.Errorf("鉴权中心返回状态码: %d", resp.StatusCode)
	}

	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		s.logger.Error("解析响应失败", elog.FieldErr(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return err
	}

	return nil
}
