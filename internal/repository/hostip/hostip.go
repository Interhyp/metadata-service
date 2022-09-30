package hostip

import (
	"errors"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"net"
)

type Impl struct {
	Logging librepo.Logging
}

func (r *Impl) ObtainLocalIp() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.To4() != nil {
				if !ip.IsLoopback() {
					return ip.To4(), nil
				}
			}
		}
	}

	return nil, errors.New("could not determine local IPv4 address, not localhost")
}
