package mtcp

import (
	"time"

	mdata "github.com/go-gost/core/metadata"
	mdutil "github.com/go-gost/core/metadata/util"
	"github.com/liukeqqs/x-master/internal/util/mux"
)

type metadata struct {
	handshakeTimeout time.Duration
	muxCfg           *mux.Config
}

func (d *mtcpDialer) parseMetadata(md mdata.Metadata) (err error) {
	d.md.handshakeTimeout = mdutil.GetDuration(md, "handshakeTimeout")

	d.md.muxCfg = &mux.Config{
		Version:           mdutil.GetInt(md, "mux.version"),
		KeepAliveInterval: mdutil.GetDuration(md, "mux.keepaliveInterval"),
		KeepAliveDisabled: mdutil.GetBool(md, "mux.keepaliveDisabled"),
		KeepAliveTimeout:  mdutil.GetDuration(md, "mux.keepaliveTimeout"),
		MaxFrameSize:      mdutil.GetInt(md, "mux.maxFrameSize"),
		MaxReceiveBuffer:  mdutil.GetInt(md, "mux.maxReceiveBuffer"),
		MaxStreamBuffer:   mdutil.GetInt(md, "mux.maxStreamBuffer"),
	}
	if d.md.muxCfg.Version == 0 {
		d.md.muxCfg.Version = 2
	}

	return
}
