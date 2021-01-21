package http

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/backend"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NewHttpBackendTransport() *BackendTransportService {
	return &BackendTransportService{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

type BackendTransportService struct {
	httpClient *http.Client
}

func (ex *BackendTransportService) Exchange(ctx flux.Context) *flux.ServeError {
	return backend.DoExchange(ctx, ex)
}

func (ex *BackendTransportService) Invoke(service flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	inurl, _ := ctx.Request().RequestURL()
	body, _ := ctx.Request().RequestBodyReader()
	newRequest, err := ex.Assemble(&service, inurl, body, ctx)
	if nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageHttpAssembleFailed,
			Internal:   err,
		}
	}
	return ex.ExecuteRequest(newRequest, service, ctx)
}

func (ex *BackendTransportService) ExecuteRequest(newRequest *http.Request, _ flux.BackendService, ctx flux.Context) (interface{}, *flux.ServeError) {
	// Header透传以及传递AttrValues
	if header, writable := ctx.Request().HeaderValues(); writable {
		newRequest.Header = header.Clone()
	} else {
		newRequest.Header = header
	}
	for k, v := range ctx.Attributes() {
		newRequest.Header.Set(k, cast.ToString(v))
	}
	resp, err := ex.httpClient.Do(newRequest)
	if nil != err {
		msg := flux.ErrorMessageHttpInvokeFailed
		if uErr, ok := err.(*url.Error); ok {
			msg = fmt.Sprintf("HTTPEX:REMOTE_ERROR:%s", uErr.Error())
		}
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayBackend,
			Message:    msg,
			Internal:   err,
		}
	}
	return resp, nil
}

func (ex *BackendTransportService) Assemble(service *flux.BackendService, inURL *url.URL, bodyReader io.ReadCloser, ctx flux.Context) (*http.Request, error) {
	inParams := service.Arguments
	newQuery := inURL.RawQuery
	// 使用可重复读的GetBody函数
	defer func() {
		_ = bodyReader.Close()
	}()
	var newBodyReader io.Reader = bodyReader
	if len(inParams) > 0 {
		// 如果Endpoint定义了参数，即表示限定参数传递
		var data string
		if values, err := _toHttpUrlValues(inParams, ctx); nil != err {
			return nil, err
		} else {
			data = values.Encode()
		}
		// GET：参数拼接到URL中；
		if http.MethodGet == service.Method {
			if newQuery == "" {
				newQuery = data
			} else {
				newQuery += "&" + data
			}
		} else {
			// 其它方法：拼接到Body中，并设置form-data/x-www-url-encoded
			newBodyReader = strings.NewReader(data)
		}
	}
	// 未定义参数，即透传Http请求：Rewrite inRequest path
	newUrl := &url.URL{
		Host:       service.RemoteHost,
		Path:       service.Interface,
		Scheme:     service.Scheme,
		Opaque:     inURL.Opaque,
		User:       inURL.User,
		RawPath:    inURL.RawPath,
		ForceQuery: inURL.ForceQuery,
		RawQuery:   newQuery,
		Fragment:   inURL.Fragment,
	}
	to := service.AttrRpcTimeout()
	timeout, err := time.ParseDuration(to)
	if err != nil {
		logger.Warnf("Illegal endpoint rpc-timeout: ", to)
		timeout = time.Second * 10
	}
	toctx, _ := context.WithTimeout(ctx.Context(), timeout)
	newRequest, err := http.NewRequestWithContext(toctx, service.Method, newUrl.String(), newBodyReader)
	if nil != err {
		return nil, fmt.Errorf("new request, method: %s, url: %s, err: %w", service.Method, newUrl, err)
	}
	// Body数据设置application/x-www-url-encoded
	if http.MethodGet != service.Method {
		newRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	newRequest.Header.Set("User-Agent", "FluxGo/Backend/v1")
	return newRequest, err
}
