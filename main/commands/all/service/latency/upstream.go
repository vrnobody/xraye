package latency

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xtls/xray-core/common/serial"
)

type Request struct {
	Auth    string  `json:"auth"`
	Uid     string  `json:"uid"`
	Ok      bool    `json:"ok"`
	Msg     string  `json:"msg"`
	Latency []int64 `json:"latency"`
	Avg     int64   `json:"avg"`
}

func (r *Request) String() string {
	return fmt.Sprintf("uid: %s, ok: %t, avg: %dms, latency: %s", r.Uid, r.Ok, r.Avg, serial.Concat(r.Latency))
}

type Response struct {
	Uid       string `json:"uid"`
	Ok        bool   `json:"ok"`
	Msg       string `json:"msg"`
	Shutdown  bool   `json:"shutdown"`
	Config    string `json:"config"`
	Cycle     int    `json:"cycle"`
	UserAgent string `json:"userAgent"`
	Timeout   int64  `json:"timeout"`
	Url       string `json:"url"`
	ExpLen    int    `json:"expLen"`
}

type Upstream struct {
	url    string
	auth   string
	client http.Client
}

func NewUpstream(url string, auth string) *Upstream {
	up := Upstream{
		url:  url,
		auth: auth,
	}
	up.client = http.Client{
		Timeout: 10 * time.Second,
	}
	return &up
}

func (up *Upstream) Post(req *Request) (*Response, error) {
	req.Auth = up.auth
	j, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := up.client.Post(up.url, "application/json", bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	r := &Response{}
	err = json.Unmarshal(body, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
