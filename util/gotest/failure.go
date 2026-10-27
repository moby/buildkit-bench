package gotest

import (
	"sort"
	"strconv"
	"strings"

	"gotest.tools/gotestsum/testjson"
)

// Failure is an unsuccessful test, benchmark attempt, or package execution.
// Name retains all matrix attributes so repeated attempts remain distinct.
type Failure struct {
	Name    string
	Package string
	Output  string
}

const diagnosticLimit = 16 * 1024

func (e *eventHandler) recordFailure(event testjson.TestEvent) {
	name := event.Test
	key := event.Package + "\x00" + name
	if event.Action == testjson.ActionOutput {
		// Benchmark failures may only have textual markers, with the event
		// attributed to the parent benchmark instead of the failed attempt.
		line := strings.TrimSpace(event.Output)
		if rest, ok := strings.CutPrefix(line, "--- FAIL: "); ok {
			name, _, _ = strings.Cut(rest, " ")
			key = event.Package + "\x00" + name
			e.failures[key] = &Failure{Name: name, Package: event.Package}
			e.activeFailure[event.Package] = key
		} else if strings.HasPrefix(line, "Benchmark") || strings.HasPrefix(line, "=== ") || strings.HasPrefix(line, "--- ") ||
			line == "FAIL" || strings.HasPrefix(line, "FAIL\t") || strings.HasPrefix(line, "exit status ") {
			delete(e.activeFailure, event.Package)
		} else if active := e.activeFailure[event.Package]; active != "" &&
			(event.Test == "" || e.failures[active].Name == event.Test || strings.HasPrefix(e.failures[active].Name, event.Test+"/")) {
			key = active
		}
		e.outputs[key] += event.Output
		if len(e.outputs[key]) > diagnosticLimit {
			const marker = "\n... output truncated ...\n"
			const head = diagnosticLimit / 2
			const tail = diagnosticLimit - head - len(marker)
			output := e.outputs[key]
			e.outputs[key] = output[:head] + marker + output[len(output)-tail:]
		}
	}
	if event.Action == testjson.ActionFail {
		// gotestsum synthesizes failures for benchmarks without terminal
		// events when their parent fails, including measured siblings.
		if event.Elapsed < 0 {
			for _, entry := range e.Benchmarks {
				for _, run := range entry.(*Benchmark).Runs {
					if strings.TrimSuffix(run.Name, "-"+strconv.Itoa(run.Threads)) == name {
						return
					}
				}
			}
		}
		if e.failures[key] == nil {
			e.failures[key] = &Failure{Name: name, Package: event.Package}
		}
	}
}

func (e *eventHandler) collectFailures() {
	for key, failure := range e.failures {
		// Parent and package failure events repeat child failures. Keep the
		// most specific records, or the package record when no child failed.
		parent := false
		for otherKey, other := range e.failures {
			if otherKey != key && other.Package == failure.Package &&
				(failure.Name == "" || strings.HasPrefix(other.Name, failure.Name+"/")) {
				parent = true
				break
			}
		}
		if !parent {
			failure.Output = e.outputs[key]
			e.Failures = append(e.Failures, *failure)
		}
	}
	sort.Slice(e.Failures, func(i, j int) bool {
		if e.Failures[i].Package != e.Failures[j].Package {
			return e.Failures[i].Package < e.Failures[j].Package
		}
		return e.Failures[i].Name < e.Failures[j].Name
	})
}
