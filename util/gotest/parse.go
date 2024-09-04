package gotest

import (
	"io"
	"log"
	"strings"
	"sync"

	"gotest.tools/gotestsum/testjson"
)

type ParseConfig struct {
	Stdout io.Reader
	Logger io.Writer
}

type ParseResult struct {
	Tests         map[string]ParseEntry
	Benchmarks    map[string]ParseEntry
	BenchmarkInfo *BenchmarkInfo
}

type ParseEntry interface {
	ID() string
	Update(event testjson.TestEvent) error
}

func Parse(config ParseConfig) (*ParseResult, *testjson.Execution, error) {
	handler := newEventHandler(config.Logger)
	ex, err := testjson.ScanTestOutput(testjson.ScanConfig{
		Stdout:  config.Stdout,
		Handler: handler,
	})
	if err != nil {
		return nil, nil, err
	}
	return &handler.ParseResult, ex, nil
}

type eventHandler struct {
	ParseResult
	log          *log.Logger
	mu           sync.Mutex
	errs         []string
	benchmarkBuf string
}

func newEventHandler(logger io.Writer) *eventHandler {
	eh := &eventHandler{
		ParseResult: ParseResult{
			BenchmarkInfo: new(BenchmarkInfo),
			Tests:         make(map[string]ParseEntry),
			Benchmarks:    make(map[string]ParseEntry),
		},
	}
	if logger != nil {
		eh.log = log.New(logger, "", log.LstdFlags)
	}
	return eh
}

func (e *eventHandler) Event(event testjson.TestEvent, _ *testjson.Execution) error {
	if event.Action != testjson.ActionOutput {
		// TODO: handle other event types (run, pass, fail, etc)
		return nil
	}
	if e.log != nil {
		log.Printf(event.Output)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if output is an incomplete benchmark. If so, buffer it and wait
	// for the next event: https://github.com/moby/buildkit-bench/issues/39
	if strings.HasPrefix(event.Output, "Benchmark") && strings.HasSuffix(event.Output, "\t") {
		e.benchmarkBuf = event.Output
		return nil
	}
	if e.benchmarkBuf != "" {
		event.Output = e.benchmarkBuf + event.Output
		e.benchmarkBuf = ""
	}

	if ok := e.BenchmarkInfo.update(event.Output); ok {
		return nil
	}
	if strings.HasPrefix(event.Output, "Benchmark") {
		return handleEvent(e.Benchmarks, event, newBenchmark)
	}
	return handleEvent(e.Tests, event, newTest)
}

func (e *eventHandler) Err(text string) error {
	e.errs = append(e.errs, text)
	return nil
}

func handleEvent(entries map[string]ParseEntry, event testjson.TestEvent, entryFunc func(testjson.TestEvent) (ParseEntry, error)) error {
	if entries == nil {
		entries = make(map[string]ParseEntry)
	}
	entry, err := entryFunc(event)
	if err != nil {
		return err
	}
	if entry != nil {
		existingEntry, ok := entries[entry.ID()]
		if !ok {
			existingEntry = entry
		} else {
			if err := existingEntry.Update(event); err != nil {
				return err
			}
		}
		entries[entry.ID()] = existingEntry
	}
	return nil
}
