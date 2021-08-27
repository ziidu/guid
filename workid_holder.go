package guid

import (
	"context"
	"fmt"
	"net"
)

// IWorkIDHolder acquire a workid for current server
type IWorkIDHolder interface {
	// WordId returns a workid for current server
	WorkId(ctx context.Context) (int, error)
}

// WorkIDFunc implements IWorkIDHolder with a function
type WorkIDFunc func(ctx context.Context) (int, error)

func (w WorkIDFunc) WorkId(ctx context.Context) (int, error) {
	return w(ctx)
}

// IpWorkIdHolder generate a workID by ipv4
var IpWorkIdHolder WorkIDFunc = func(ctx context.Context) (int, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}
	return int(ip[0]) ^ int(ip[1]) ^ int(ip[2]) ^ int(ip[3]), nil
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}
