package unix

import (
	"net"

	"github.com/go-gost/core/listener"
	"github.com/go-gost/core/logger"
	md "github.com/go-gost/core/metadata"
	admission "github.com/liukeqqs/x-master/admission/wrapper"
	climiter "github.com/liukeqqs/x-master/limiter/conn/wrapper"
	limiter "github.com/liukeqqs/x-master/limiter/traffic/wrapper"
	metrics "github.com/liukeqqs/x-master/metrics/wrapper"
	"github.com/liukeqqs/x-master/registry"
	stats "github.com/liukeqqs/x-master/stats/wrapper"
)

func init() {
	registry.ListenerRegistry().Register("unix", NewListener)
}

type unixListener struct {
	ln      net.Listener
	logger  logger.Logger
	md      metadata
	options listener.Options
}

func NewListener(opts ...listener.Option) listener.Listener {
	options := listener.Options{}
	for _, opt := range opts {
		opt(&options)
	}
	return &unixListener{
		logger:  options.Logger,
		options: options,
	}
}

func (l *unixListener) Init(md md.Metadata) (err error) {
	if err = l.parseMetadata(md); err != nil {
		return
	}

	ln, err := net.Listen("unix", l.options.Addr)
	if err != nil {
		return
	}

	ln = metrics.WrapListener(l.options.Service, ln)
	ln = stats.WrapListener(ln, l.options.Stats)
	ln = admission.WrapListener(l.options.Admission, ln)
	ln = limiter.WrapListener(l.options.TrafficLimiter, ln)
	ln = climiter.WrapListener(l.options.ConnLimiter, ln)
	l.ln = ln

	return
}

func (l *unixListener) Accept() (conn net.Conn, err error) {
	return l.ln.Accept()
}

func (l *unixListener) Addr() net.Addr {
	return l.ln.Addr()
}

func (l *unixListener) Close() error {
	return l.ln.Close()
}
