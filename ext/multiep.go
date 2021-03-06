package ext

import (
	"github.com/bytepowered/flux"
	"sync"
)

var (
	endpoints = new(sync.Map)
)

func SelectEndpoint(key string) (*flux.MultiEndpoint, bool) {
	ep, ok := endpoints.Load(key)
	if ok {
		return ep.(*flux.MultiEndpoint), true
	}
	return nil, false
}

func RegisterEndpoint(key string, endpoint *flux.Endpoint) *flux.MultiEndpoint {
	mve := flux.NewMultiEndpoint(endpoint)
	endpoints.Store(key, mve)
	return mve
}

func LoadEndpoints() map[string]*flux.MultiEndpoint {
	out := make(map[string]*flux.MultiEndpoint, 32)
	endpoints.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(*flux.MultiEndpoint)
		return true
	})
	return out
}
