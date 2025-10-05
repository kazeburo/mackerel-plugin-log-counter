package main

import (
	"fmt"

	"github.com/kazeburo/followparser"
	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/mackerelio/golib/pluginutil"
)

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
	if u.opt.LogArchiveDir != "" {
		fp.ArchiveDir = u.opt.LogArchiveDir
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
