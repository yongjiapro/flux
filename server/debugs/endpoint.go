package debugs

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/server"
	"net/http"
)

const (
	queryKeyApplication  = "application"
	queryKeyProtocol     = "protocol"
	queryKeyHttpPattern  = "http-pattern"
	queryKeyHttpPattern0 = "httpPattern"
	queryKeyHttpPattern1 = "httppattern"
	queryKeyInterface    = "interface"
)

type EndpointFilter func(ep *server.BindEndpoint) bool

// 支持以下过滤条件
var endpointQueryKeys = []string{queryKeyApplication, queryKeyProtocol,
	queryKeyHttpPattern, queryKeyHttpPattern0, queryKeyHttpPattern1,
	queryKeyInterface,
}

var (
	endpointFilterFactories = make(map[string]func(string) EndpointFilter)
)

func init() {
	endpointFilterFactories[queryKeyApplication] = func(query string) EndpointFilter {
		return func(ep *server.BindEndpoint) bool {
			return queryMatch(query, ep.RandomVersion().Application)
		}
	}
	endpointFilterFactories[queryKeyProtocol] = func(query string) EndpointFilter {
		return func(ep *server.BindEndpoint) bool {
			proto := ep.RandomVersion().Service.RpcProto
			return queryMatch(query, proto)
		}
	}
	httpPatternFilter := func(query string) EndpointFilter {
		return func(ep *server.BindEndpoint) bool {
			return queryMatch(query, ep.RandomVersion().HttpPattern)
		}
	}
	endpointFilterFactories[queryKeyHttpPattern] = httpPatternFilter
	endpointFilterFactories[queryKeyHttpPattern0] = httpPatternFilter
	endpointFilterFactories[queryKeyHttpPattern1] = httpPatternFilter

	endpointFilterFactories[queryKeyInterface] = func(query string) EndpointFilter {
		return func(ep *server.BindEndpoint) bool {
			return queryMatch(query, ep.RandomVersion().Service.Interface)
		}
	}
}

// NewDebugQueryEndpointHandler Endpoint查询
func NewDebugQueryEndpointHandler(datamap map[string]*server.BindEndpoint) http.HandlerFunc {
	serializer := ext.LoadSerializer(ext.TypeNameSerializerJson)
	return newSerializableHttpHandler(serializer, func(request *http.Request) interface{} {
		return queryEndpoints(datamap, request)
	})
}

func queryEndpoints(data map[string]*server.BindEndpoint, request *http.Request) interface{} {
	filters := make([]EndpointFilter, 0)
	query := request.URL.Query()
	for _, key := range endpointQueryKeys {
		if query := query.Get(key); "" != query {
			if f, ok := endpointFilterFactories[key]; ok {
				filters = append(filters, f(query))
			}
		}
	}
	if len(filters) == 0 {
		m := make(map[string]map[string]*flux.Endpoint, 16)
		for k, v := range data {
			m[k] = v.ToSerializable()
		}
		return m
	}
	return queryWithEndpointFilters(data, filters...)
}

func queryWithEndpointFilters(data map[string]*server.BindEndpoint, filters ...EndpointFilter) []map[string]*flux.Endpoint {
	items := make([]map[string]*flux.Endpoint, 0, 16)
DataLoop:
	for _, v := range data {
		for _, filter := range filters {
			// 任意Filter返回True
			if filter(v) {
				items = append(items, v.ToSerializable())
				continue DataLoop
			}
		}
	}
	return items
}
