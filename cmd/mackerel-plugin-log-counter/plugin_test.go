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
		t.Fatalf("failed to create log file: %v", err)
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
		assert.Equal(t, m, map[string]float64{
			"pattern1": 300,
			"pattern2": 300,
		}, "match metrics")
	}
}

func TestLogCounterPlugin_RotateAndArchive(t *testing.T) {
	tmpdir := t.TempDir()
	archiveDir := filepath.Join(tmpdir, "archive")
	if err := os.Mkdir(archiveDir, 0755); err != nil {
		t.Fatalf("failed to create archive dir: %v", err)
	}

	logFileName := filepath.Join(tmpdir, "log")
	fh, err := os.Create(logFileName)
	if err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	patterns := []patternReg{
		{name: "pattern1", reg: regexp.MustCompile(`error`)},
		{name: "pattern2", reg: regexp.MustCompile(`warning`)},
	}

	// initial run to create tracking files
	{
		opt := Opt{
			Prefix:        "test",
			patternRegs:   patterns,
			LogFile:       logFileName,
			LogArchiveDir: archiveDir,
			PerSec:        true,
		}
		plugin := LogCounterPlugin{opt: opt}

		_, err := plugin.FetchMetrics()
		assert.NoError(t, err)
	}

	// write some lines to the original log file, then rotate (move to archive)
	for i := 0; i < 10; i++ {
		fh.WriteString(fmt.Sprintf("warning msg %08d\n", i))
	}
	for i := 0; i < 10; i++ {
		fh.WriteString(fmt.Sprintf("error msg %08d\n", i))
	}
	fh.Sync()

	// rotate: move the file to archive
	if err := fh.Close(); err != nil {
		t.Fatalf("failed to close log file: %v", err)
	}
	archived := filepath.Join(archiveDir, "log.1")
	if err := os.Rename(logFileName, archived); err != nil {
		t.Fatalf("failed to rotate log file: %v", err)
	}

	// create a new log file with the same name and write additional lines
	fh2, err := os.Create(logFileName)
	if err != nil {
		t.Fatalf("failed to create new log file: %v", err)
	}
	for i := 0; i < 5; i++ {
		fh2.WriteString(fmt.Sprintf("warning msg %08d\n", i))
	}
	for i := 0; i < 5; i++ {
		fh2.WriteString(fmt.Sprintf("error msg %08d\n", i))
	}
	fh2.Sync()

	// allow followparser to measure a non-zero duration
	time.Sleep(time.Second)

	{
		opt := Opt{
			Prefix:        "test",
			patternRegs:   patterns,
			LogFile:       logFileName,
			LogArchiveDir: archiveDir,
			PerSec:        true,
		}
		plugin := LogCounterPlugin{opt: opt}

		m, err := plugin.FetchMetrics()
		assert.NoError(t, err)
		// expect counts from both archived (10 each) and new file (5 each)
		assert.Equal(t, map[string]float64{
			"pattern1": 15,
			"pattern2": 15,
		}, m, "counts should include archived and new entries")
	}
}
