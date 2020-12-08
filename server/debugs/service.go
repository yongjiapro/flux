package debugs

import (
	"github.com/bytepowered/flux/ext"
	"net/http"
)

const (
	queryKeyServiceId  = "serviceid"
	queryKeyServiceId0 = "service-id"
	queryKeyServiceId1 = "serviceId"
)

var (
	serviceQueryKeys = []string{queryKeyServiceId, queryKeyServiceId0, queryKeyServiceId1}
)

// NewDebugQueryServiceHandler Service查询
func NewDebugQueryServiceHandler() http.HandlerFunc {
	serializer := ext.LoadSerializer(ext.TypeNameSerializerJson)
	return newSerializableHttpHandler(serializer, func(request *http.Request) interface{} {
		query := request.URL.Query()
		for _, key := range serviceQueryKeys {
			if id := query.Get(key); "" != id {
				service, ok := ext.LoadBackendService(id)
				if ok {
					return service
				} else {
					return map[string]string{
						"status":     "failed",
						"message":    "service not found",
						"service-id": id,
					}
				}
			}
		}
		return map[string]string{
			"status":  "failed",
			"message": "param is required: serviceId",
		}
	})
}
