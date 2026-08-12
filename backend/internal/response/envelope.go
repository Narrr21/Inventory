// Package response writes the single JSON envelope every endpoint shares.
//
// The shape is deliberately flat (api-contract.md v6): a client reads
// `success` and `code` the same way whether the request worked or not, and
// never has to reach into a nested object that only exists on failure.
//
//	{ "success": true,  "code": 200, "data": {...}, "meta": {...} }
//	{ "success": false, "code": 404, "errorCode": "NOT_FOUND", "msg": "..." }
//
// `code` always mirrors the HTTP status, so it's derived from the status
// argument here rather than passed separately — the two can't drift.
package response

import (
	"encoding/json"
	"net/http"
)

type successEnvelope struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type successWithErrorsEnvelope struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type errorEnvelope struct {
	Success bool `json:"success"`
	Code    int  `json:"code"`
	// ErrorCode is the stable machine-readable reason. It carries information
	// the HTTP status alone doesn't: 409 means either PROJECT_IN_USE or
	// ITEM_TYPE_IN_USE, and clients branch on which.
	ErrorCode string            `json:"errorCode"`
	Msg       string            `json:"msg"`
	Fields    map[string]string `json:"fields,omitempty"`
	// Details carries structured context for errors where the client needs
	// more than prose — currently the blocking item count on 409, so the UI
	// can say "12 items still use this" without a follow-up query.
	Details map[string]interface{} `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func OK(w http.ResponseWriter, status int, data interface{}, meta interface{}) {
	writeJSON(w, status, successEnvelope{Success: true, Code: status, Data: data, Meta: meta})
}

func OKWithErrors(w http.ResponseWriter, status int, data interface{}, errors interface{}) {
	writeJSON(w, status, successWithErrorsEnvelope{Success: true, Code: status, Data: data, Errors: errors})
}

// Err writes a failure envelope. fields may be nil for errors that aren't
// per-field validation failures.
func Err(w http.ResponseWriter, status int, errorCode string, msg string, fields map[string]string) {
	writeJSON(w, status, errorEnvelope{
		Success:   false,
		Code:      status,
		ErrorCode: errorCode,
		Msg:       msg,
		Fields:    fields,
	})
}

// ErrDetails writes a failure envelope with structured details instead of
// per-field messages — used by the 409 delete-blocked responses.
func ErrDetails(w http.ResponseWriter, status int, errorCode string, msg string, details map[string]interface{}) {
	writeJSON(w, status, errorEnvelope{
		Success:   false,
		Code:      status,
		ErrorCode: errorCode,
		Msg:       msg,
		Details:   details,
	})
}
