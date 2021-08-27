package model

import (
	"encoding/json"
	"net/http"
)

const (
	TimeBackErrCode = 5000

	NotFoundErrCode = 4004
)

type Response struct {
	Id   int64 `json:"id"`
	Code int   `json:"code"`
}

func ResponseOk(id int64) Response {
	return Response{Id: id, Code: http.StatusOK}
}

func ResponseErr(code int) Response {
	return Response{Code: code}
}

func (resp Response) Encoding() ([]byte, error) {
	return json.Marshal(&resp)
}
