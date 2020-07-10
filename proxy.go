package tunnel

import (
	"io"
	"net"

	"github.com/koding/logging"
	"github.com/koding/tunnel/proto"
)

// ProxyFunc is responsible for forwarding a remote connection to local server and writing the response back.
type ProxyFunc func(remote net.Conn, msg *proto.ControlMessage)

var (
	// DefaultProxyFuncs holds global default proxy functions for all transport protocols.
	DefaultProxyFuncs = ProxyFuncs{
		HTTP: new(HTTPProxy).Proxy,
		TCP:  new(TCPProxy).Proxy,
		WS:   new(HTTPProxy).Proxy,
	}
	// DefaultProxy is a ProxyFunc that uses DefaultProxyFuncs.
	DefaultProxy = Proxy(ProxyFuncs{})
)

// ProxyFuncs is a collection of ProxyFunc.
type ProxyFuncs struct {
	// HTTP is custom implementation of HTTP proxing.
	HTTP ProxyFunc
	// TCP is custom implementation of TCP proxing.
	TCP ProxyFunc
	// WS is custom implementation of web socket proxing.
	WS ProxyFunc
}

// Proxy returns a ProxyFunc that uses custom function if provided, otherwise falls back to DefaultProxyFuncs.
func Proxy(p ProxyFuncs) ProxyFunc {
	return func(remote net.Conn, msg *proto.ControlMessage) {
		var f ProxyFunc
		switch msg.Protocol {
		case proto.HTTP:
			f = DefaultProxyFuncs.HTTP
			if p.HTTP != nil {
				f = p.HTTP
			}
		case proto.TCP:
			f = DefaultProxyFuncs.TCP
			if p.TCP != nil {
				f = p.TCP
			}
		case proto.WS:
			f = DefaultProxyFuncs.WS
			if p.WS != nil {
				f = p.WS
			}
		}

		if f == nil {
			logging.Error("Could not determine proxy function for %v", msg)
			remote.Close()
		}

		f(remote, msg)
	}
}

// Join copies data between local and remote connections.
// It reads from one connection and writes to the other.
// It's a building block for ProxyFunc implementations.
func Join(local, remote net.Conn, log logging.Logger) {
	done := make(chan interface{}, 2)

	transfer := func(side string, dst, src net.Conn) {
		defer func() { done <- nil }()

		log.Debug("proxing %s -> %s", src.RemoteAddr(), dst.RemoteAddr())
		n, _ := io.Copy(dst, src)
		// n, err :=
		// if err != nil {
		// 	log.Error("%s: copy error: %s", side, err)
		// }
		log.Debug("done proxing %s -> %s: %d bytes", src.RemoteAddr(), dst.RemoteAddr(), n)
	}

	go transfer("remote to local", local, remote)
	go transfer("local to remote", remote, local)

	<-done
	local.Close()
	remote.Close()
}
