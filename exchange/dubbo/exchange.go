package dubbo

import (
	"context"
	"errors"
	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	"github.com/apache/dubbo-go/common/constant"
	_ "github.com/apache/dubbo-go/common/proxy/proxy_factory"
	dubbogo "github.com/apache/dubbo-go/config"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/registry/nacos"
	_ "github.com/apache/dubbo-go/registry/protocol"
	_ "github.com/apache/dubbo-go/registry/zookeeper"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"sync"
)

const (
	ConfigName = "DubboExchange"
)

var (
	ErrInvalidHeaders = errors.New("DUBBO_RPC:INVALID_HEADERS")
	ErrInvalidStatus  = errors.New("DUBBO_RPC:INVALID_STATUS")
	ErrMessageInvoke  = "DUBBO_RPC:INVOKE"
	ErrMessageDecode  = "DUBBO_RPC:DECODE"
)

// 集成DubboRPC框架的Exchange
type exchange struct {
	config       flux.Config
	referenceMap map[string]*dubbogo.ReferenceConfig
	referenceMu  sync.RWMutex
}

func NewDubboExchange() flux.Exchange {
	return &exchange{
		referenceMap: make(map[string]*dubbogo.ReferenceConfig),
	}
}

func (ex *exchange) Init(config flux.Config) error {
	logger.Infof("Dubbo Exchange initializing")
	ex.config = config
	if 0 == len(ex.config) {
		return errors.New("dubbo-exchange config not found")
	} else {
		return nil
	}
}

func (ex *exchange) Name() string {
	return ConfigName
}

func (ex *exchange) Exchange(ctx flux.Context) *flux.InvokeError {
	endpoint := ctx.Endpoint()
	resp, err := ex.Invoke(&endpoint, ctx)
	if err != nil {
		return err
	}
	// decode response
	if code, headers, body, err := decode(resp); nil == err {
		ctx.ResponseWriter().SetStatusCode(code).SetHeaders(headers).SetBody(body)
		return nil
	} else {
		return &flux.InvokeError{
			StatusCode: flux.StatusServerError,
			Message:    ErrMessageDecode,
			Internal:   err,
		}
	}
}

func (ex *exchange) Invoke(target *flux.Endpoint, reqCtx flux.Context) (interface{}, *flux.InvokeError) {
	types, args := assemble(target.Arguments)
	reference := ex.lookup(target)
	goctx := context.Background()
	if nil != reqCtx {
		goctx = context.WithValue(context.Background(), constant.AttachmentKey, pkg.ToStringKVMap(reqCtx.AttrValues()))
	}
	if resp, err := reference.GetRPCService().(*dubbogo.GenericService).
		Invoke(goctx, []interface{}{target.UpstreamMethod, types, args}); err != nil {
		logger.Infof("Dubbo rpc error, service: %s, method: %s, err: %s", target.UpstreamUri, target.UpstreamMethod, err)
		return nil, &flux.InvokeError{
			StatusCode: flux.StatusBadGateway,
			Message:    ErrMessageInvoke,
			Internal:   err,
		}
	} else {
		return resp, nil
	}
}

func (ex *exchange) lookup(endpoint *flux.Endpoint) *dubbogo.ReferenceConfig {
	ex.referenceMu.Lock()
	defer ex.referenceMu.Unlock()
	interfaceName := endpoint.UpstreamUri
	if ref, ok := ex.referenceMap[interfaceName]; ok {
		return ref
	} else {
		newRef := newReference(endpoint, ex.config)
		ex.referenceMap[interfaceName] = newRef
		return newRef
	}
}