package v5

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/gosocks5"
	ctxvalue "github.com/liukeqqs/x-master/ctx"
	"github.com/liukeqqs/x-master/internal/net/udp"
	"github.com/liukeqqs/x-master/internal/util/socks"
	"github.com/liukeqqs/x-master/stats"
	stats_wrapper "github.com/liukeqqs/x-master/stats/wrapper"
)

func (h *socks5Handler) handleUDP(ctx context.Context, conn net.Conn, log logger.Logger) error {
	log = log.WithFields(map[string]any{
		"cmd": "udp",
	})

	if !h.md.enableUDP {
		reply := gosocks5.NewReply(gosocks5.NotAllowed, nil)
		log.Trace(reply)
		log.Error("socks5: UDP relay is disabled")
		return reply.Write(conn)
	}

	cc, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Error(err)
		reply := gosocks5.NewReply(gosocks5.Failure, nil)
		log.Trace(reply)
		reply.Write(conn)
		return err
	}
	defer cc.Close()

	saddr := gosocks5.Addr{}
	saddr.ParseFrom(cc.LocalAddr().String())
	saddr.Type = 0
	saddr.Host, _, _ = net.SplitHostPort(conn.LocalAddr().String()) // replace the IP to the out-going interface's
	reply := gosocks5.NewReply(gosocks5.Succeeded, &saddr)
	log.Trace(reply)
	if err := reply.Write(conn); err != nil {
		log.Error(err)
		return err
	}

	log = log.WithFields(map[string]any{
		"bind": fmt.Sprintf("%s/%s", cc.LocalAddr(), cc.LocalAddr().Network()),
	})
	log.Debugf("bind on %s OK", cc.LocalAddr())

	// obtain a udp connection
	c, err := h.router.Dial(ctx, "udp", "") // UDP association
	if err != nil {
		log.Error(err)
		return err
	}
	defer c.Close()

	pc, ok := c.(net.PacketConn)
	if !ok {
		err := errors.New("socks5: wrong connection type")
		log.Error(err)
		return err
	}

	var lc net.PacketConn = cc
	clientID := ctxvalue.ClientIDFromContext(ctx)
	if h.options.Observer != nil {
		pstats := h.stats.Stats(string(clientID))
		pstats.Add(stats.KindTotalConns, 1)
		pstats.Add(stats.KindCurrentConns, 1)
		defer pstats.Add(stats.KindCurrentConns, -1)
		lc = stats_wrapper.WrapPacketConn(lc, pstats)
	}

	r := udp.NewRelay(socks.UDPConn(lc, h.md.udpBufferSize), pc).
		WithBypass(h.options.Bypass).
		WithLogger(log)
	r.SetBufferSize(h.md.udpBufferSize)

	go r.Run(ctx)

	t := time.Now()
	log.Debugf("%s <-> %s", conn.RemoteAddr(), cc.LocalAddr())
	io.Copy(io.Discard, conn)
	log.WithFields(map[string]any{"duration": time.Since(t)}).
		Debugf("%s >-< %s", conn.RemoteAddr(), cc.LocalAddr())

	return nil
}
