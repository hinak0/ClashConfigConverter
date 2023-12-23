package proto

type RawConfig struct {
	Port               int      `yaml:"port"`
	SocksPort          int      `yaml:"socks-port,omitempty"`
	RedirPort          int      `yaml:"redir-port,omitempty"`
	TProxyPort         int      `yaml:"tproxy-port,omitempty"`
	MixedPort          int      `yaml:"mixed-port,omitempty"`
	Authentication     []string `yaml:"authentication,omitempty"`
	AllowLan           bool     `yaml:"allow-lan,omitempty"`
	BindAddress        string   `yaml:"bind-address,omitempty"`
	Mode               any      `yaml:"mode,omitempty"`
	LogLevel           any      `yaml:"log-level,omitempty"`
	IPv6               bool     `yaml:"ipv6,omitempty"`
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
	Proxy         []map[string]interface{}  `yaml:"proxies,omitempty"`
	ProxyGroup    []ProxyGroup              `yaml:"proxy-groups,omitempty"`
	Rule          []string                  `yaml:"rules,omitempty"`
}

type RawDNS struct {
	Enable            bool              `yaml:"enable,omitempty"`
	IPv6              *bool             `yaml:"ipv6,omitempty"`
	UseHosts          bool              `yaml:"use-hosts,omitempty"`
	NameServer        []string          `yaml:"nameserver,omitempty"`
	Fallback          []string          `yaml:"fallback,omitempty"`
	FallbackFilter    any               `yaml:"fallback-filter,omitempty"`
	Listen            string            `yaml:"listen,omitempty"`
	EnhancedMode      any               `yaml:"enhanced-mode,omitempty"`
	FakeIPRange       string            `yaml:"fake-ip-range,omitempty"`
	FakeIPFilter      []string          `yaml:"fake-ip-filter,omitempty"`
	DefaultNameserver []string          `yaml:"default-nameserver,omitempty"`
	NameServerPolicy  map[string]string `yaml:"nameserver-policy,omitempty"`
	SearchDomains     []string          `yaml:"search-domains,omitempty"`
}

type ProxyGroup struct {
	Name      string   `yaml:"name"`
	Type      string   `yaml:"type"`
	URL       string   `yaml:"url,omitempty"`
	Interval  int      `yaml:"interval,omitempty"`
	Tolerance int      `yaml:"tolerance,omitempty"`
	Proxies   []string `yaml:"proxies"`
}
