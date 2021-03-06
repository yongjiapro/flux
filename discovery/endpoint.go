package discovery

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/remoting"
)

var (
	invalidHttpEndpointEvent = flux.HttpEndpointEvent{}
)

type CompatibleEndpoint struct {
	flux.Endpoint
	// Deprecated
	Authorize bool `json:"authorize"` // 此端点是否需要授权
}

func NewEndpointEvent(bytes []byte, etype remoting.EventType, node string) (fxEvt flux.HttpEndpointEvent, ok bool) {
	// Check json text
	size := len(bytes)
	if size < len("{\"k\":0}") || (bytes[0] != '[' && bytes[size-1] != '}') {
		logger.Warnw("DISCOVERY:ENDPOINT:ILLEGAL_JSONSIZE", "data", string(bytes), "node", node)
		return invalidHttpEndpointEvent, false
	}
	comp := CompatibleEndpoint{}
	if err := ext.JSONUnmarshal(bytes, &comp); nil != err {
		logger.Warnw("DISCOVERY:ENDPOINT:ILLEGAL_JSONFORMAT",
			"event-type", etype, "data", string(bytes), "error", err, "node", node)
		return invalidHttpEndpointEvent, false
	}
	// 检查有效性
	if !comp.IsValid() {
		logger.Warnw("DISCOVERY:ENDPOINT:INVALID_VALUES", "data", string(bytes), "node", node)
		return invalidHttpEndpointEvent, false
	}
	// 兼容旧结构
	if len(comp.Attributes) == 0 {
		comp.Attributes = []flux.Attribute{
			{Name: flux.EndpointAttrTagAuthorize, Value: comp.Authorize},
		}
	}
	EnsureServiceAttrs(&comp.Service)
	EnsureServiceAttrs(&comp.Permission)

	event := flux.HttpEndpointEvent{Endpoint: comp.Endpoint}
	switch etype {
	case remoting.EventTypeNodeAdd:
		event.EventType = flux.EventTypeAdded
	case remoting.EventTypeNodeDelete:
		event.EventType = flux.EventTypeRemoved
	case remoting.EventTypeNodeUpdate:
		event.EventType = flux.EventTypeUpdated
	default:
		return invalidHttpEndpointEvent, false
	}
	return event, true
}
