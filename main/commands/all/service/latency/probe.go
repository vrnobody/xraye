package latency

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/xtls/xray-core/common/errors"
	clog "github.com/xtls/xray-core/common/log"
	v2net "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
)

func startHeadlessInstance(cfg string, keep bool) (*core.Instance, error) {
	config, err := core.LoadConfig("json", bytes.NewBufferString(cfg))
	if err != nil {
		return nil, err
	}

	if !keep {
		// remove inbounds section of config.json
		config.Inbound = []*core.InboundHandlerConfig{}
	}

	instance, err := core.New(config)
	if err != nil {
		return nil, err
	}

	// disable logging
	clog.ReplaceWithSeverityLogger(clog.Severity_Unknown)

	if err := instance.Start(); err != nil {
		return nil, err
	}
	return instance, nil
}

func doGetRequest(httpClient *http.Client, req *http.Request, resp *Response) (int64, error) {
	startTime := time.Now()
	response, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	if response.Body == nil {
		return 0, errors.New("html body is empty!")
	}
	defer response.Body.Close()
	size := 0
	buff := make([]byte, 4096)
	for resp.ExpLen < 1 || size < resp.ExpLen {
		n, err := response.Body.Read(buff)
		if err == io.EOF {
			// download completed!
			break
		} else if err != nil {
			return 0, err
		} else if n > 0 {
			size = size + n
		} else {
			// download completed!
			break
		}
	}
	endTime := time.Now()
	if resp.ExpLen > 0 && size < resp.ExpLen {
		return 0, errors.New("html body size is smaller than expected")
	}
	return endTime.Sub(startTime).Milliseconds(), nil
}

func calcAvg(old int64, new int64) int64 {
	if old < 1 {
		return new
	}
	if new < 1 {
		return old
	}
	// old * 60% + new * 40%
	return (old*6 + new*4) / 10
}

func createGetRequest(resp *Response) (*http.Request, error) {
	req, err := http.NewRequest("GET", resp.Url, nil)
	if err != nil {
		return nil, err
	}
	if len(resp.UserAgent) > 0 {
		req.Header.Set("User-Agent", resp.UserAgent)
	}
	return req, nil
}

func probe(wlog *Logger, resp *Response) (*Request, error) {
	xray, err := startHeadlessInstance(resp.Config, resp.Keep)
	if err != nil {
		return nil, err
	}
	defer xray.Close()

	httpClient, err := createHttpClient(xray, resp)
	if err != nil {
		return nil, err
	}

	req, err := createGetRequest(resp)
	if err != nil {
		return nil, err
	}

	avg := int64(0)
	durs := []int64{}
	for i := 0; i < resp.Cycle; i++ {
		dur, err := doGetRequest(httpClient, req, resp)
		if err != nil {
			logIf(wlog.Warn, resp.Uid, fmt.Sprintf("round: %d, error: %s", i+1, err))
			continue
		}

		if dur > 0 {
			durs = append(durs, dur)
			avg = calcAvg(avg, dur)
			tail := fmt.Sprintf("round: %d, avg: %dms, latency: %s", i+1, avg, serial.Concat(durs))
			logIf(wlog.Debug, resp.Uid, tail)
		}
	}

	r := &Request{
		Ok:      true,
		Uid:     resp.Uid,
		Latency: durs,
		Avg:     int64(avg),
	}
	return r, nil
}

func createHttpClient(xray *core.Instance, resp *Response) (*http.Client, error) {
	httpTransport := http.Transport{
		Proxy: func(*http.Request) (*url.URL, error) {
			return nil, nil
		},
		DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			dest, err := v2net.ParseDestination(network + ":" + addr)
			if err != nil {
				return nil, err
			}
			conn, err := core.Dial(ctx, xray, dest)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		TLSHandshakeTimeout: time.Second * 6,
	}
	httpClient := &http.Client{
		Timeout:   time.Millisecond * time.Duration(resp.Timeout),
		Transport: &httpTransport,
		Jar:       nil,
	}
	return httpClient, nil
}
