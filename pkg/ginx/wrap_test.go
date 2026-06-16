package ginx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockBusinessError struct {
	code int
	msg  string
}

func (m mockBusinessError) Error() string {
	return m.msg
}

func (m mockBusinessError) GetCode() int {
	return m.code
}

func (m mockBusinessError) GetMsg() string {
	return m.msg
}

func TestHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		err          error
		systemResult Result
		wantStatus   int
		wantBody     string
	}{
		{
			name: "业务错误拦截",
			err:  mockBusinessError{code: 400001, msg: "唯一标识冲突"},
			systemResult: Result{
				Code: 502001,
				Msg:  "系统错误",
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"code":400001,"msg":"唯一标识冲突","data":null}`,
		},
		{
			name: "业务错误包装拦截",
			err:  errors.Join(errors.New("模型创建失败"), mockBusinessError{code: 400001, msg: "唯一标识冲突"}),
			systemResult: Result{
				Code: 502001,
				Msg:  "系统错误",
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"code":400001,"msg":"唯一标识冲突","data":null}`,
		},
		{
			name: "普通系统错误 fallback",
			err:  errors.New("数据库连接断开"),
			systemResult: Result{
				Code: 502001,
				Msg:  "系统错误",
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"code":502001,"msg":"系统错误","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			handleError(c, tt.err, tt.systemResult)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}
