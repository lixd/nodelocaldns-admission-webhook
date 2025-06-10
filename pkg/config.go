package pkg

type Config struct {
	KubeDNS  string
	LocalDNS string
	Domain   string
}

const (
	DefaultKubeDNS  = "10.96.0.10"
	DefaultLocalDNS = "169.254.20.10"
	DefaultDomain   = "cluster.local"
)

func NewDNSConfig(kubedns, localdns, domain string) Config {
	if kubedns == "" {
		kubedns = DefaultKubeDNS
	}
	if localdns == "" {
		localdns = DefaultLocalDNS
	}
	if domain == "" {
		domain = DefaultDomain
	}
	return Config{
		KubeDNS:  kubedns,
		LocalDNS: localdns,
		Domain:   domain,
	}
}
