package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogCounterPlugin_FetchMetrics(t *testing.T) {
	tmpdir := t.TempDir()
	logFileName := filepath.Join(tmpdir, "log")
	fh, err := os.Create(logFileName)
	if err != nil {
		t.Error(err)
	}

	patterns := []patternReg{
		{name: "pattern1", reg: regexp.MustCompile(`error`)},
		{name: "pattern2", reg: regexp.MustCompile(`warning`)},
	}
	{
		opt := Opt{
			Prefix:      "test",
			patternRegs: patterns,
			LogFile:     logFileName,
			PerSec:      true,
		}
		plugin := LogCounterPlugin{opt: opt}

		_, err := plugin.FetchMetrics()
		assert.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("warning msg %08d\n", i)
		fh.WriteString(msg)
	}
	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("error msg %08d\n", i)
		fh.WriteString(msg)
	}

	time.Sleep(time.Second)

	{
		opt := Opt{
			Prefix:      "test",
			patternRegs: patterns,
			LogFile:     logFileName,
			PerSec:      true,
		}
		plugin := LogCounterPlugin{opt: opt}

		m, err := plugin.FetchMetrics()
		assert.NoError(t, err)
		assert.Equal(t, m, map[string]float64{
			"pattern1": 10,
			"pattern2": 10,
		}, "match metrics")
	}

	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("warning msg %08d\n", i)
		fh.WriteString(msg)
	}
	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("error msg %08d\n", i)
		fh.WriteString(msg)
	}
	time.Sleep(time.Second)
	{
		opt := Opt{
			Prefix:      "test",
			patternRegs: patterns,
			LogFile:     logFileName,
			PerSec:      false,
		}
		plugin := LogCounterPlugin{opt: opt}

		m, err := plugin.FetchMetrics()
		assert.NoError(t, err)
		assert.Equal(t, &m, &map[string]float64{
			"pattern1": 300,
			"pattern2": 300,
		}, "match meterics")
	}
}
