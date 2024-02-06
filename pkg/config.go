package pkg

type Config struct {
	KubeDNS  string
	LocalDNS string
}

const (
	DefaultKubeDNS  = "10.96.0.10"
	DefaultLocalDNS = "169.254.20.10"
)

func NewDNSConfig(kubedns, localdns string) Config {
	if kubedns == "" {
		kubedns = DefaultKubeDNS
	}
	if localdns == "" {
		localdns = DefaultLocalDNS
	}
	return Config{
		KubeDNS:  kubedns,
		LocalDNS: localdns,
	}
}
