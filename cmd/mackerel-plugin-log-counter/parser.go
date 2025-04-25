package main

import "bytes"

type Parser struct {
	opt        Opt
	mapCounter map[string]float64
	duration   float64
}

func NewParser(opt Opt) *Parser {
	m := map[string]float64{}
	for _, pr := range opt.patternRegs {
		m[pr.name] = float64(0)
	}
	return &Parser{
		opt:        opt,
		mapCounter: m,
	}
}

func (p *Parser) Parse(b []byte) error {
	if p.opt.filterByte != nil && !bytes.Contains(b, *p.opt.filterByte) {
		return nil
	}
	if p.opt.ignoreByte != nil && bytes.Contains(b, *p.opt.ignoreByte) {
		return nil
	}
	for _, pr := range p.opt.patternRegs {
		if pr.reg.Match(b) {
			p.mapCounter[pr.name]++
		}
	}
	return nil
}

func (p *Parser) Finish(duration float64) {
	p.duration = duration
}

func (p *Parser) GetResult() map[string]float64 {
	m := map[string]float64{}
	if p.duration == 0 {
		// first running
		return m
	}
	for _, pr := range p.opt.patternRegs {
		m[pr.name] = p.mapCounter[pr.name]
		if p.opt.PerSec {
			m[pr.name] = m[pr.name] / p.duration
		} else {
			m[pr.name] = (m[pr.name] / p.duration) * 60
		}
	}
	return m
}
