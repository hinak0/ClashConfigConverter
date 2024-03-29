package proto

type RawConfig struct {
	Port               int      `yaml:"port"`
	SocksPort          int      `yaml:"socks-port,omitempty"`
	RedirPort          int      `yaml:"redir-port,omitempty"`
	TProxyPort         int      `yaml:"tproxy-port,omitempty"`
	MixedPort          int      `yaml:"mixed-port,omitempty"`
	Authentication     []string `yaml:"authentication,omitempty"`
	AllowLan           bool     `yaml:"allow-lan"`
	BindAddress        string   `yaml:"bind-address,omitempty"`
	Mode               string   `yaml:"mode,omitempty"`
	LogLevel           string   `yaml:"log-level,omitempty"`
	IPv6               bool     `yaml:"ipv6"`
	ExternalController string   `yaml:"external-controller,omitempty"`
	ExternalUI         string   `yaml:"external-ui,omitempty"`
	Secret             string   `yaml:"secret,omitempty"`
	Interface          string   `yaml:"interface-name,omitempty"`
	RoutingMark        int      `yaml:"routing-mark,omitempty"`
	Tunnels            []any    `yaml:"tunnels,omitempty"`

	ProxyProvider map[string]map[string]any `yaml:"proxy-providers,omitempty"`
	Hosts         map[string]string         `yaml:"hosts,omitempty"`
	Inbounds      []any                     `yaml:"inbounds,omitempty"`
	DNS           RawDNS                    `yaml:"dns,omitempty"`
	Experimental  any                       `yaml:"experimental,omitempty"`
	Profile       any                       `yaml:"profile,omitempty"`
	Proxy         []Proxy                   `yaml:"proxies,omitempty"`
	ProxyGroup    []ProxyGroup              `yaml:"proxy-groups,omitempty"`
	Rule          []string                  `yaml:"rules,omitempty"`

	Others map[string]interface{} `yaml:",inline"`
}

type RawDNS struct {
	Enable            bool              `yaml:"enable"`
	UseHosts          bool              `yaml:"use-hosts"`
	NameServer        []any             `yaml:"nameserver"`
	Fallback          []any             `yaml:"fallback"`
	FallbackFilter    any               `yaml:"fallback-filter,omitempty"`
	Listen            string            `yaml:"listen,omitempty"`
	EnhancedMode      string            `yaml:"enhanced-mode,omitempty"`
	FakeIPRange       string            `yaml:"fake-ip-range,omitempty"`
	FakeIPFilter      []string          `yaml:"fake-ip-filter,omitempty"`
	DefaultNameserver []string          `yaml:"default-nameserver,omitempty"`
	NameServerPolicy  map[string]string `yaml:"nameserver-policy,omitempty"`
	SearchDomains     []string          `yaml:"search-domains,omitempty"`

	Others map[string]interface{} `yaml:",inline"`
}

type ProxyGroup struct {
	Name    string                 `yaml:"name"`
	Type    string                 `yaml:"type"`
	Proxies []string               `yaml:"proxies"`
	Others  map[string]interface{} `yaml:",inline"`
}

type Proxy struct {
	Name      string `yaml:"name"`
	Server    string `yaml:"server"`
	Port      int    `yaml:"port"`
	UdpEnable bool   `yaml:"udp"`

	Others map[string]interface{} `yaml:",inline"`
}
