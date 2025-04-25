package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"

	"github.com/jessevdk/go-flags"
	mp "github.com/mackerelio/go-mackerel-plugin"
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
		fmt.Fprint(os.Stderr, "Specify --pattern and --key-name\n")
		return 1
	}
	if len(opt.KeyNames) != len(opt.Patterns) {
		fmt.Fprint(os.Stderr, "The number of --pattern and --key-name must be the same\n")
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
