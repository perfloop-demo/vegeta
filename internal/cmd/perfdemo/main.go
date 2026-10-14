// Command perfdemo runs one opt-in performance demonstration scenario from
// package perfdemo over deterministic results. It prints a digest of the
// output, so two runs can be compared, and the elapsed time.
//
//	go run ./internal/cmd/perfdemo -scenario report -n 4000
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/tsenart/vegeta/v12/internal/perfdemo"
)

func main() {
	scenario := flag.String("scenario", "report", "Scenario to run: report, tally or percentiles")
	n := flag.Int("n", 4000, "Number of deterministic results")
	flag.Parse()

	rs := perfdemo.Results(*n)
	start := time.Now()
	var out string
	switch *scenario {
	case "report":
		out = perfdemo.Report(rs)
	case "tally":
		out = fmt.Sprint(perfdemo.Tally(rs, "GET", []byte("ok")))
	case "percentiles":
		out = fmt.Sprint(perfdemo.Percentiles(rs, []float64{0.5, 0.9, 0.95, 0.99}))
	default:
		log.Fatalf("unknown scenario %q", *scenario)
	}
	elapsed := time.Since(start)

	fmt.Printf("scenario=%s n=%d sha256=%x elapsed=%s\n", *scenario, *n, sha256.Sum256([]byte(out)), elapsed)
}
