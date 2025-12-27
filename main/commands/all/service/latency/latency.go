package latency

import (
	"fmt"
	"io"

	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/main/commands/base"
	"github.com/xtls/xray-core/main/confloader"
)

var CmdLatency = &base.Command{
	CustomFlags: true,
	UsageLine:   "{{.Exec}} service latency [--url=https://www.google.com/] [stdin:] [config.json]",
	Short:       "Test latency for one config file.",
	Long: `
Test latency for one config file.

Arguments:
    --url
        Web page for latency testing. Default https://www.google.com/.

    --timeout
        Timeout in milliseconds. Default 10000.

    --cycle
        Perform n rounds of latency test. Default 1.

    --keep
        Keep config.json as-is. Default false means to remove inbounds section of config.json.

    --explen
        Expected html text length. Default zero means to download the entire web page.

    --useragent
        Custom User-Agent http header.

Examples:

    {{.Exec}} service latency config.json
    `,
	Run: executeCmdLatency,
}

func executeCmdLatency(cmd *base.Command, args []string) {

	var url string
	var timeout int64
	var cycle, explen int
	var keep bool
	var useragent string

	cmd.Flag.StringVar(&url, "url", "https://www.google.com/", "")
	cmd.Flag.Int64Var(&timeout, "timeout", 10000, "")
	cmd.Flag.IntVar(&cycle, "cycle", 1, "")
	cmd.Flag.IntVar(&explen, "explen", 0, "")
	cmd.Flag.BoolVar(&keep, "keep", false, "")
	cmd.Flag.StringVar(&useragent, "useragent", "", "")
	cmd.Flag.Parse(args)

	if cmd.Flag.NArg() < 1 {
		base.Fatalf("empty input list")
	}

	reader, err := confloader.LoadConfig(cmd.Flag.Arg(0))
	if err != nil {
		base.Fatalf("failed to load config: %s", err)
	}
	b, err := io.ReadAll(reader)
	if err != nil {
		base.Fatalf("failed to load config: %s", err)
	}

	options := &Response{
		Uid:       "",
		Ok:        true,
		Config:    string(b),
		Keep:      keep,
		Cycle:     cycle,
		UserAgent: useragent,
		Timeout:   timeout,
		Url:       url,
		ExpLen:    explen,
	}

	log := &Logger{
		loglevel: Loglevel_Debug,
	}
	log.Info("probing: ", options.Url)
	r, err := probe(log, options)
	if err != nil {
		log.Warn("error: ", err)
		return
	}

	log.Info(fmt.Sprintf("average: %dms, latency: %s", r.Avg, serial.Concat(r.Latency)))
}
