package latency

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/main/commands/base"
)

var CmdProber = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} service prober [--workers=1] [--url=http://localhost:4001/] [--auth=passwrod]",
	Short:       "Batch latency tester.",
	Long: `
Batch latency tester.

There is a tutorial on how to setup a upstream server in GitHub wiki pages.

Arguments:

    -u, --url <http://host:port/path/>
        Upstream server URL. Default http://localhost:4001/.

    -w, --workers <num>
        Concurrent latency test workers number. Default 1.

    --auth <password>
        Upstream server password.

Examples:

    {{.Exec}} service prober -w 1 -u http://localhost:4001 --auth="123456"
    `,
	Run: executeRun,
}

func executeRun(cmd *base.Command, args []string) {
	var url, auth string
	var wnum int
	cmd.Flag.StringVar(&url, "u", "http://localhost:4001/", "")
	cmd.Flag.StringVar(&url, "url", "http://localhost:4001/", "")
	cmd.Flag.StringVar(&auth, "auth", "", "")
	cmd.Flag.IntVar(&wnum, "w", 1, "")
	cmd.Flag.IntVar(&wnum, "workers", 1, "")
	cmd.Flag.Parse(args)

	mlog := &Logger{
		loglevel: Loglevel_Info,
	}
	upstream := &Upstream{
		auth: auth,
		url:  url,
	}
	mlog.Info("prober starts")
	var ctx, cancel = context.WithCancel(context.Background())
	wg, err := startWorkers(ctx, upstream, wnum)
	if err != nil {
		mlog.Error(err)
		cancel()
		return
	}
	go watchSigTerm(mlog, cancel)
	wg.Wait()
	mlog.Info("prober stopped")
}

func startWorkers(ctx context.Context, upstream *Upstream, wnum int) (*sync.WaitGroup, error) {
	if wnum < 1 {
		return nil, errors.New("workers number must larger than 0")
	}
	var wg sync.WaitGroup
	for i := 1; i <= wnum; i++ {
		wg.Go(func() {
			wlog := &Logger{
				loglevel: Loglevel_Info,
				tag:      fmt.Sprintf("[w%d]", i),
			}
			worker(ctx, wlog, upstream)
		})
	}
	return &wg, nil
}

func watchSigTerm(mlog *Logger, cancel context.CancelFunc) {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	<-osSignals
	mlog.Warn("detect stop signal")
	cancel()
}

func doWork(wlog *Logger, resp *Response) (req *Request) {
	timeout := &Request{
		Uid: resp.Uid,
		Ok:  true,
		Avg: 0,
	}

	defer func() {
		if err := recover(); err != nil {
			wlog.Warn("uid: ", resp.Uid, ", panic: ", err)
			req = timeout
		}
	}()

	wlog.Info("uid: ", resp.Uid, ", probe: ", resp.Url)
	req, err := probe(wlog, resp)
	if err != nil {
		wlog.Warn("uid: ", resp.Uid, ", error: ", err)
		return timeout
	}
	wlog.Info(req)
	return req
}

func worker(ctx context.Context, wlog *Logger, upstream *Upstream) {
	wlog.Info("starts")
	defer wlog.Info("stopped")
	req := &Request{}

	for {
		select {
		case <-ctx.Done():
			if req.Ok {
				// ignore error
				upstream.Post(req)
			}
			return
		default:
			resp, err := upstream.Post(req)
			if err != nil {
				wlog.Warn("upstream error: ", err)
			} else if resp.Shutdown {
				wlog.Warn("receive shutdown signal")
				return
			} else if !resp.Ok {
				wlog.Warn("upstream error: ", resp.Msg)
			} else {
				req = doWork(wlog, resp)
			}
			if err != nil || !resp.Ok || !req.Ok {
				time.Sleep(2 * time.Second)
			}
		}
	}
}
