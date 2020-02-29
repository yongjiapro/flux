package flux

type (
	EndpointEventType int
)

// 路由元数据事件类型
const (
	EndpointEventAdded = iota
	EndpointEventUpdated
	EndpointEventRemoved
)

const (
	// 自动查找数据源
	ScopeAuto = "AUTO"
	// 获取Http Attributes的单个参数
	ScopeAttr = "ATTR"
	// 获取Http Attributes的Map结果
	ScopeAttrs = "ATTRS"
	// 只从Form表单参数参数列表中读取
	ScopeForm = "FORM"
	// 只从Header参数中读取
	ScopeHeader = "HEADER"
	// 只从Query和Form表单参数参数列表中读取
	ScopeParam = "PARAM"
	// 从动态Path参数中获取
	ScopePath = "PATH"
	// 从Query参数中获取
	ScopeQuery = "QUERY"
)

const (
	// 原始参数类型：int,long...
	ArgumentTypePrimitive = "PRIMITIVE"
	// 复杂参数类型：POJO
	ArgumentTypeComplex = "COMPLEX"
)

// Support protocols
const (
	ProtocolDubbo = "DUBBO"
	ProtocolGRPC  = "GRPC"
	ProtocolHttp  = "HTTP"
	ProtocolEcho  = "ECHO"
)

type (
	// Argument 定义Endpoint的参数结构元数据
	Argument struct {
		TypeClass   string     `json:"typeClass"`    // 参数类型
		TypeGeneric []string   `json:"typeGeneric"` // 泛型类型
		ArgName     string     `json:"argName"`     // 参数名称
		ArgType     string     `json:"argType"`     // 参数结构类型：字段、POJO
		ArgValue    Valuer     `json:"-"`           // 参数值
		HttpName    string     `json:"httpName"`    // 映射Http的参数名
		HttpScope   string     `json:"httpScope"`   // 映射Http参数值域
		Fields      []Argument `json:"fields"`      // POJO类型时的子结构
	}

	// Endpoint 定义前端Http请求与后端RPC服务的端点元数据
	Endpoint struct {
		Application    string     `json:"application"`    // 所属应用名
		Version        string     `json:"version"`        // 定义的版本号
		Protocol       string     `json:"protocol"`       // 支持的协议
		RpcGroup       string     `json:"rpcGroup"`       // rpc接口分组
		RpcVersion     string     `json:"rpcVersion"`     // rpc接口版本
		RpcTimeout     string     `json:"rpcTimeout"`     // RPC调用超时
		RpcRetries     string     `json:"rpcRetries"`     // RPC调用重试
		Authorize      bool       `json:"authorize"`      // 此端点是否需要授权
		UpstreamHost   string     `json:"upstreamHost"`   // 定义Upstream侧的Host
		UpstreamUri    string     `json:"upstreamUri"`    // 定义Upstream侧的URL
		UpstreamMethod string     `json:"upstreamMethod"` // 定义Upstream侧的方法
		HttpPattern    string     `json:"httpPattern"`    // 映射Http侧的UriPattern
		HttpMethod     string     `json:"httpMethod"`     // 映射Http侧的Method
		Arguments      []Argument `json:"arguments"`      // 参数结构
	}

	// EndpointEvent  定义从注册中心接收到的Endpoint数据变更
	EndpointEvent struct {
		Type        EndpointEventType
		HttpMethod  string `json:"method"`
		HttpPattern string `json:"pattern"`
		Endpoint    Endpoint
	}
)

/// Value

func NewWrapValue(v interface{}) Valuer {
	return &ValueWrapper{
		value: v,
	}
}

type ValueWrapper struct {
	value interface{}
}

func (v *ValueWrapper) Value() interface{} {
	return v.value
}

func (v *ValueWrapper) SetValue(value interface{}) {
	v.value = value
}
