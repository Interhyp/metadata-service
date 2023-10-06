package repository

import "net"

// HostIP interacts with the local network interfaces.
type HostIP interface {
	IsHostIP() bool

	Setup() error

	// ObtainLocalIp gets the first non-localhost ipv4 address from your interfaces.
	//
	// In a k8s deployment, that'll be the pod ip.
	ObtainLocalIp() (net.IP, error)
}
