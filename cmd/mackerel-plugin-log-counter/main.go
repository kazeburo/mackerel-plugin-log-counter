package main

import (
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

type Opt struct {
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

type LogCounterPlugin struct {
	opt Opt
}

func (u LogCounterPlugin) GraphDefinition() map[string]mp.Graphs {
	metrics := make([]mp.Metrics, 0)
	for _, pr := range u.opt.patternRegs {
		metrics = append(metrics, mp.Metrics{
			Name:    pr.name,
			Label:   pr.name,
			Diff:    false,
			Stacked: false,
		})
	}
	comment := "(per minute)"
	if u.opt.PerSec {
		comment = "(per second)"
	}
	return map[string]mp.Graphs{
		"": {
			Label:   fmt.Sprintf("LogCounter %s %s", u.opt.Prefix, comment),
			Unit:    mp.UnitFloat,
			Metrics: metrics,
		},
	}
}

func (u LogCounterPlugin) FetchMetrics() (map[string]float64, error) {
	p := NewParser(u.opt)
	fp := &followparser.Parser{
		WorkDir:  pluginutil.PluginWorkDir(),
		Callback: p,
		Silent:   !u.opt.Verbose,
	}
	_, err := fp.Parse(
		fmt.Sprintf("%s-mp-log-counter", u.opt.Prefix),
		u.opt.LogFile,
	)
	if err != nil {
		return nil, err
	}
	return p.GetResult(), nil
}

func (u LogCounterPlugin) MetricKeyPrefix() string {
	return u.opt.Prefix
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opt := Opt{}
	psr := flags.NewParser(&opt, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opt.Version {
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

	if len(opt.KeyNames) == 0 {
		fmt.Fprint(os.Stderr, "Specify --pattern and --key-name\n", err)
		return 1
	}
	if len(opt.KeyNames) != len(opt.Patterns) {
		fmt.Fprint(os.Stderr, "The number of --pattern and --key-name must be the same\n", err)
		return 1
	}

	patterns := make([]patternReg, 0)
	for i, k := range opt.KeyNames {
		reg, err := regexp.Compile(opt.Patterns[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		patterns = append(patterns, patternReg{
			reg:  reg,
			name: k,
		})
	}
	opt.patternRegs = patterns

	if opt.Filter != "" {
		b := []byte(opt.Filter)
		opt.filterByte = &b
	}
	if opt.Ignore != "" {
		b := []byte(opt.Ignore)
		opt.ignoreByte = &b
	}

	u := LogCounterPlugin{
		opt: opt,
	}
	plugin := mp.NewMackerelPlugin(u)
	plugin.Run()
	return 0
}
