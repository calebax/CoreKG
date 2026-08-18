package dtouser

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ErrorResponse struct {
	Code    uint32 `json:"code"`    // 错误码，
	Message string `json:"message"` // 错误信息，
}

type RequestPasswordResetCodeResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

type ForgotPasswordResponse struct {
	apiobj.BaseResponse
	Response struct{}
}
