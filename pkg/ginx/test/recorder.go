package test

import (
	"encoding/json"
	"github.com/ecodeclub/ekit/net/httpx/httptestx"
	"net/http/httptest"
)

func NewJSONResponseRecorder[T any]() *httptestx.JSONResponseRecorder[Result[T]] {
	return httptestx.NewJSONResponseRecorder[Result[T]]()
}

type JSONResponseRecorder[T any] struct {
	*httptest.ResponseRecorder
}

func (r JSONResponseRecorder[T]) Scan() (T, error) {
	var t T
	err := json.NewDecoder(r.Body).Decode(&t)
	return t, err
}

func (r JSONResponseRecorder[T]) MustScan() T {
	t, err := r.Scan()
	if err != nil {
		panic(err)
	}
	return t
}
