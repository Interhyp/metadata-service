package hostip

import (
	"context"
	"errors"
	"github.com/Interhyp/metadata-service/internal/acorn/repository"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	librepo "github.com/StephanHCB/go-backend-service-common/acorns/repository"
	"net"
)

type Impl struct {
	Logging librepo.Logging
}

func New(
	logging librepo.Logging,
) repository.HostIP {
	return &Impl{
		Logging: logging,
	}
}

func (r *Impl) IsHostIP() bool {
	return true
}

func (r *Impl) Setup() error {
	ctx := auzerolog.AddLoggerToCtx(context.Background())

	ip, err := r.ObtainLocalIp()
	if err != nil {
		r.Logging.Logger().Ctx(ctx).Error().WithErr(err).Print("failed to obtain local ip address. BAILING OUT")
		return err
	}

	r.Logging.Logger().Ctx(ctx).Info().Printf("non-trivial ipv4 address is %s", ip.String())

	return nil
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
