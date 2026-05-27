package helpers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/edkin/url-shortener/pkg/logger"
)

type ErrorResp struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ErrResp(code, message string) ErrorResp {
	return ErrorResp{Error: ErrorBody{Code: code, Message: message}}
}

func InternalErr() ErrorResp {
	return ErrResp("INTERNAL_ERROR", "internal server error")
}

func WriteJSON(ctx context.Context, w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.FromContext(ctx).Error("response encode failed", logger.Error(err))
	}
}
