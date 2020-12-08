package debugs

import (
	"github.com/bytepowered/flux"
	"net/http"
	"strings"
)

func newSerializableHttpHandler(serializer flux.Serializer, queryHandler func(request *http.Request) interface{}) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		data := queryHandler(request)
		if data, err := serializer.Marshal(data); nil != err {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
		} else {
			writer.WriteHeader(http.StatusOK)
			writer.Header().Set("Content-Type", "application/json;charset=UTF-8")
			_, _ = writer.Write(data)
		}
	}
}

func queryMatch(input, expected string) bool {
	input, expected = strings.ToLower(input), strings.ToLower(expected)
	return input == expected || strings.Contains(expected, input)
}
