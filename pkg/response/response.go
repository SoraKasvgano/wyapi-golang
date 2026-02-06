package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Code    int         `json:"code"`
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Msg     string      `json:"msg,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func Write(w http.ResponseWriter, httpStatus int, code int, msg string, data interface{}) {
	resp := APIResponse{
		Code:    code,
		Status:  code,
		Success: code >= 200 && code < 300,
		Msg:     msg,
		Message: msg,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(resp)
}

func Success(w http.ResponseWriter, data interface{}, msg string) {
	Write(w, http.StatusOK, http.StatusOK, msg, data)
}

func Error(w http.ResponseWriter, httpStatus int, msg string) {
	if httpStatus == 0 {
		httpStatus = http.StatusBadRequest
	}
	Write(w, httpStatus, httpStatus, msg, nil)
}
