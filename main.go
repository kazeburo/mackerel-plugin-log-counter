package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"runtime"

	"github.com/jessevdk/go-flags"
	"github.com/kazeburo/followparser"
	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/mackerelio/golib/pluginutil"
)

// version by Makefile
var version string

type patternReg struct {
	reg  *regexp.Regexp
	name string
}

type cmdOpts struct {
	Version     bool     `short:"v" long:"version" description:"Show version"`
	Filter      string   `long:"filter" description:"filter string used before check pattern."`
	Ignore      string   `long:"ignore" description:"ignore string used before check pattern."`
	Patterns    []string `short:"p" long:"pattern" required:"true" description:"Regexp pattern to search for."`
	KeyNames    []string `short:"k" long:"key-name" required:"true" description:"Key name for pattern"`
	Prefix      string   `long:"prefix" required:"true" description:"Metric key prefix"`
	LogFile     string   `long:"log-file" default:"/var/log/messages" description:"Path to log file" required:"true"`
	PerSec      bool     `long:"per-second" description:"calcurate per-seconds count. default per minute count"`
	Verbose     bool     `long:"verbose" description:"display infomational logs"`
	patternRegs []patternReg
	filterByte  *[]byte
	ignoreByte  *[]byte
}

type parser struct {
	opts     cmdOpts
	cnt      map[string]float64
	duration float64
}

func NewParser(opts cmdOpts) *parser {
	m := map[string]float64{}
	for _, pr := range opts.patternRegs {
		m[pr.name] = float64(0)
	}
	return &parser{
		opts: opts,
		cnt:  m,
	}
}

func (p *parser) Parse(b []byte) error {
	if p.opts.filterByte != nil && !bytes.Contains(b, *p.opts.filterByte) {
		return nil
	}
	if p.opts.ignoreByte != nil && bytes.Contains(b, *p.opts.ignoreByte) {
		return nil
	}
	for _, pr := range p.opts.patternRegs {
		if pr.reg.Match(b) {
			p.cnt[pr.name]++
		}
	}
	return nil
}

func (p *parser) Finish(duration float64) {
	p.duration = duration
}

func (p *parser) GetResult() map[string]float64 {
	m := map[string]float64{}
	if p.duration == 0 {
		// first running
		return m
	}
	for _, pr := range p.opts.patternRegs {
		m[pr.name] = p.cnt[pr.name]
		if p.opts.PerSec {
			m[pr.name] = m[pr.name] / p.duration
		} else {
			m[pr.name] = (m[pr.name] / p.duration) * 60
		}
	}
	return m
}

type LogCounterPlugin struct {
	opts cmdOpts
}

func (u LogCounterPlugin) GraphDefinition() map[string]mp.Graphs {
	metrics := make([]mp.Metrics, 0)
	for _, pr := range u.opts.patternRegs {
		metrics = append(metrics, mp.Metrics{
			Name:    pr.name,
			Label:   pr.name,
			Diff:    false,
			Stacked: false,
		})
	}
	comment := "(per minute)"
	if u.opts.PerSec {
		comment = "(per second)"
	}
	return map[string]mp.Graphs{
		"": {
			Label:   fmt.Sprintf("LogCounter %s %s", u.opts.Prefix, comment),
			Unit:    mp.UnitFloat,
			Metrics: metrics,
		},
	}
}

func (u LogCounterPlugin) FetchMetrics() (map[string]float64, error) {
	p := NewParser(u.opts)
	fp := &followparser.Parser{
		WorkDir:  pluginutil.PluginWorkDir(),
		Callback: p,
		Silent:   !u.opts.Verbose,
	}
	_, err := fp.Parse(
		fmt.Sprintf("%s-mp-log-counter", u.opts.Prefix),
		u.opts.LogFile,
	)
	if err != nil {
		return nil, err
	}
	return p.GetResult(), nil
}

func (u LogCounterPlugin) MetricKeyPrefix() string {
	return u.opts.Prefix
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opts := cmdOpts{}
	psr := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opts.Version {
		fmt.Printf(`%s %s
Compiler: %s %s
`,
			os.Args[0],
			version,
			runtime.Compiler,
			runtime.Version())
		return 0
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	if len(opts.KeyNames) == 0 {
		fmt.Fprint(os.Stderr, "Specify --pattern and --key-name\n", err)
		return 1
	}
	if len(opts.KeyNames) != len(opts.Patterns) {
		fmt.Fprint(os.Stderr, "The number of --pattern and --key-name must be the same\n", err)
		return 1
	}

	patterns := make([]patternReg, 0)
	for i, k := range opts.KeyNames {
		reg, err := regexp.Compile(opts.Patterns[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		patterns = append(patterns, patternReg{
			reg:  reg,
			name: k,
		})
	}
	opts.patternRegs = patterns

	if opts.Filter != "" {
		b := []byte(opts.Filter)
		opts.filterByte = &b
	}
	if opts.Ignore != "" {
		b := []byte(opts.Ignore)
		opts.ignoreByte = &b
	}

	u := LogCounterPlugin{opts}
	plugin := mp.NewMackerelPlugin(u)
	plugin.Run()
	return 0
}
