package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func IsZeroOfUnderlyingType(v interface{}) bool {
	return v == nil || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

type emptyStruct struct{}

type data struct {
	Records interface{} `json:"records,omitempty"`
	Record  interface{} `json:"record,omitempty"`
}

func (i *data) SetData(result interface{}) data {
	isNil := IsZeroOfUnderlyingType(result)
	if isSlice := reflect.ValueOf(result).Kind() == reflect.Slice; isSlice {
		if isNil {
			result = []emptyStruct{}
		}
		i.Records = result
	} else {
		if isNil {
			result = emptyStruct{}
		}
		i.Record = result
	}
	return *i
}

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type metaResponse struct {
	// CorrelationID is the response correlation_id
	RequestId string `json:"request_id"`
	// Code is the response code
	Code int `json:"code"`
	// Message is the response message
	Message string `json:"message"`
	//Time is the response message
	Time string `json:"time"`
	// Pagination of the pagination response
	Pagination *Pagination `json:"pagination,omitempty"`
}

type responseHttp struct {
	// Meta is the API response information
	Meta metaResponse `json:"meta"`
	// Data is our data
	Data data `json:"data"`
	// Errors is the response message
	Errors interface{} `json:"errors,omitempty"`
}

const TraceIDContextKey = "REQUEST_ID"

var (
	Loc, _        = time.LoadLocation("Asia/Ho_Chi_Minh")
	LayoutDefault = "2006-01-02 15:04:05"
)

func TraceIdentifierMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL != nil && strings.Contains(r.URL.RawQuery, ";") {
			fmt.Println(fmt.Sprintf("http: URL query contains semicolon - %s", r.URL))
		}

		traceId := r.Header.Get(TraceIDContextKey)
		if traceId == "" {
			traceId = uuid.NewString()
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, TraceIDContextKey, traceId) // nolint
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetTraceIdentifier(ctx context.Context) string {
	traceId := ctx.Value(TraceIDContextKey)
	if traceId == nil {
		return uuid.NewString()
	}
	return fmt.Sprint(traceId)
}

type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m Message) Error() string {
	return m.Message
}

func setHttpResponse(ctx context.Context, msg Message, result interface{}, paging *Pagination, err error) interface{} {
	dt := data{}
	return responseHttp{
		Meta: metaResponse{
			RequestId:  GetTraceIdentifier(ctx),
			Code:       msg.Code,
			Message:    msg.Message,
			Time:       time.Now().In(Loc).Format(LayoutDefault),
			Pagination: paging,
		},
		Data:   dt.SetData(result),
		Errors: err,
	}
}

func SetHttpResponse(ctx context.Context, msg Message, result interface{}, paging *Pagination) interface{} {
	return setHttpResponse(ctx, msg, result, paging, nil)
}

func SetDefaultResponse(ctx context.Context, msg Message) interface{} {
	return setHttpResponse(ctx, msg, nil, nil, nil)
}

func ResponseWriter(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
	return
}

func EncodeError(ctx context.Context, err error, w http.ResponseWriter) {
	msgResponse := Message{Code: 500, Message: "Internal Server Error"}
	var fieldError validator.ValidationErrors
	var message Message
	switch {
	case errors.As(err, &fieldError):
		msgResponse = Message{Code: 422, Message: err.Error()}
	case errors.As(err, &message):
		msgResponse = message
	}
	ResponseWriter(w, http.StatusInternalServerError, SetDefaultResponse(ctx, msgResponse))
}

type encodeError interface {
	error() error
}

func GetHttpResponse(resp interface{}) *responseHttp {
	if result, ok := resp.(responseHttp); ok {
		return &result
	}
	return nil
}

func EncodeResponseHTTP(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	if err, ok := resp.(encodeError); ok && err.error() != nil {
		EncodeError(ctx, err.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	result := GetHttpResponse(resp)
	w.WriteHeader(result.Meta.Code)
	return json.NewEncoder(w).Encode(resp)
}
