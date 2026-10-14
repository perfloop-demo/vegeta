// Package perfdemo holds opt-in performance demonstration fixtures.
//
// Nothing in the vegeta library or CLI imports this package. Each scenario
// reads the deterministic Results from Results, so its tests and benchmarks
// repeat exactly. The scenarios are demonstration fixtures, not
// recommendations for how vegeta should process results.
package perfdemo

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// Results returns n deterministic results. Sequence numbers, status codes,
// latencies and bodies follow fixed cycles, so equal n gives equal results.
func Results(n int) []vegeta.Result {
	methods := []string{"GET", "POST", "PUT"}
	codes := []uint16{200, 200, 200, 404, 500}
	bodies := [][]byte{[]byte("ok"), []byte("not found"), []byte("error")}
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	rs := make([]vegeta.Result, n)
	for i := range rs {
		code := codes[i%len(codes)]
		body := bodies[0]
		switch code {
		case 404:
			body = bodies[1]
		case 500:
			body = bodies[2]
		}
		rs[i] = vegeta.Result{
			Attack:    "perfdemo",
			Seq:       uint64(i),
			Code:      code,
			Timestamp: start.Add(time.Duration(i) * time.Millisecond),
			Latency:   time.Duration(1+(i*7919)%997) * time.Microsecond,
			BytesIn:   uint64(len(body)),
			Body:      body,
			Method:    methods[i%len(methods)],
			URL:       fmt.Sprintf("http://localhost:8080/items/%d", i%50),
		}
	}
	return rs
}

// Report renders one line per result: sequence, method, URL, status code
// and latency. This is the allocation-heavy scenario.
func Report(rs []vegeta.Result) string {
	out := ""
	for _, r := range rs {
		out += fmt.Sprintf("%d %s %s %d %s\n", r.Seq, r.Method, r.URL, r.Code, r.Latency)
	}
	return out
}

// Tally counts results whose method matches method in any letter case and
// whose body equals body. The per-result normalization and conversion look
// like avoidable work. The control benchmark measures the same count without
// them, so a claimed gain can be checked against it.
func Tally(rs []vegeta.Result, method string, body []byte) int {
	method = strings.ToUpper(method)
	n := 0
	for _, r := range rs {
		if strings.ToUpper(r.Method) == method && string(r.Body) == string(body) {
			n++
		}
	}
	return n
}

// Percentiles returns the nearest-rank latency for each quantile in ps, in
// the same order. Each quantile p is in [0, 1]. It returns nil for no
// results.
func Percentiles(rs []vegeta.Result, ps []float64) []time.Duration {
	if len(rs) == 0 {
		return nil
	}
	out := make([]time.Duration, len(ps))
	for i, p := range ps {
		ls := make([]time.Duration, len(rs))
		for j, r := range rs {
			ls[j] = r.Latency
		}
		sort.Slice(ls, func(a, b int) bool { return ls[a] < ls[b] })
		rank := int(math.Ceil(p*float64(len(ls)))) - 1
		if rank < 0 {
			rank = 0
		}
		if rank >= len(ls) {
			rank = len(ls) - 1
		}
		out[i] = ls[rank]
	}
	return out
}
