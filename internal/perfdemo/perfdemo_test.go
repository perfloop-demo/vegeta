package perfdemo

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestResultsDeterministic(t *testing.T) {
	a, b := Results(25), Results(25)
	for i := range a {
		if !a[i].Equal(b[i]) {
			t.Fatalf("result %d differs between calls: %+v != %+v", i, a[i], b[i])
		}
	}
	if got, want := a[3].Code, uint16(404); got != want {
		t.Fatalf("result 3 code = %d, want %d", got, want)
	}
	if got, want := string(a[4].Body), "error"; got != want {
		t.Fatalf("result 4 body = %q, want %q", got, want)
	}
}

func TestReport(t *testing.T) {
	got := Report(Results(2))
	want := "0 GET http://localhost:8080/items/0 200 1µs\n" +
		"1 POST http://localhost:8080/items/1 200 941µs\n"
	if got != want {
		t.Fatalf("Report() =\n%q\nwant\n%q", got, want)
	}
	if got := Report(nil); got != "" {
		t.Fatalf("Report(nil) = %q, want empty", got)
	}
}

func TestTally(t *testing.T) {
	rs := []vegeta.Result{
		{Method: "get", Body: []byte("ok")},
		{Method: "GET", Body: []byte("ok")},
		{Method: "Get", Body: []byte("ok!")},
		{Method: "POST", Body: []byte("ok")},
		{Method: "gEt", Body: []byte("ok")},
	}
	if got, want := Tally(rs, "Get", []byte("ok")), 3; got != want {
		t.Fatalf("Tally(Get, ok) = %d, want %d", got, want)
	}
	if got, want := Tally(rs, "post", []byte("ok")), 1; got != want {
		t.Fatalf("Tally(post, ok) = %d, want %d", got, want)
	}
	if got, want := Tally(rs, "GET", []byte("missing")), 0; got != want {
		t.Fatalf("Tally(GET, missing) = %d, want %d", got, want)
	}
}

func TestPercentiles(t *testing.T) {
	var rs []vegeta.Result
	for _, ms := range []int{5, 1, 4, 2, 3} {
		rs = append(rs, vegeta.Result{Latency: time.Duration(ms) * time.Millisecond})
	}
	got := Percentiles(rs, []float64{0.9, 0.4, 0})
	want := []time.Duration{5 * time.Millisecond, 2 * time.Millisecond, 1 * time.Millisecond}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Percentiles() = %v, want %v", got, want)
	}
	if got := Percentiles(nil, []float64{0.5}); got != nil {
		t.Fatalf("Percentiles(nil) = %v, want nil", got)
	}
}

func BenchmarkReport(b *testing.B) {
	for _, n := range []int{1000, 4000} {
		rs := Results(n)
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				Report(rs)
			}
		})
	}
}

// tallyControl is the negative control for Tally. It counts the same results
// for inputs that are already upper case, with no normalization or string
// conversion.
func tallyControl(rs []vegeta.Result, method string, body []byte) int {
	n := 0
	for _, r := range rs {
		if r.Method == method && bytes.Equal(r.Body, body) {
			n++
		}
	}
	return n
}

func BenchmarkTally(b *testing.B) {
	rs := Results(10000)
	body := []byte("ok")
	if got, want := Tally(rs, "GET", body), tallyControl(rs, "GET", body); got != want {
		b.Fatalf("Tally = %d, control = %d", got, want)
	}
	b.Run("tally", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			Tally(rs, "GET", body)
		}
	})
	b.Run("control", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tallyControl(rs, "GET", body)
		}
	})
}

func BenchmarkPercentiles(b *testing.B) {
	rs := Results(10000)
	ps := []float64{0.5, 0.9, 0.95, 0.99}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Percentiles(rs, ps)
	}
}
