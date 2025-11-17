// go run main.go -admin http://127.0.0.1:9901 -filter rq_
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Envoy /stats?format=json shape.
type envoyStats struct {
	Stats []struct {
		Name       string      `json:"name"`
		Value      interface{} `json:"value"`      // number or string depending on Envoy version
		Type       string      `json:"type"`       // "GAUGE" | "COUNTER" | ...
		Tags       interface{} `json:"tags"`       // ignore for now
		Used       bool        `json:"used"`       // may be present
		Histograms interface{} `json:"histograms"` // ignore for now
	} `json:"stats"`
}

func mkUrl(models []string) string {
	filter := fmt.Sprintf("(%s)", strings.Join(models, "|"))
	return fmt.Sprintf("/stats?filter=cluster.%s_*.*upstream_rq_total&format=json&type=Counters", filter)
}

func main() {
	admin := flag.String("admin", "http://127.0.0.1:9901", "Envoy admin base URL")
	filter := flag.String("filter", "rq_", "Substring or regex (if -regex) to include (e.g. rq_|downstream_rq_|upstream_rq_)")
	useRegex := flag.Bool("regex", false, "Interpret -filter as regex")
	models := []string{"iris2", "blah"}
	flag.Parse()

	url := *admin + mkUrl(models)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	c := &http.Client{Timeout: 5 * time.Second}

	res, err := c.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		fmt.Fprintf(os.Stderr, "admin returned %d: %s\n", res.StatusCode, string(b))
		os.Exit(1)
	}

	var payload envoyStats
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		fmt.Fprintf(os.Stderr, "decode failed: %v\n", err)
		os.Exit(1)
	}

	var match func(string) bool
	if *useRegex {
		re, err := regexp.Compile(*filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bad regex: %v\n", err)
			os.Exit(1)
		}
		match = re.MatchString
	} else {
		sub := *filter
		match = func(s string) bool { return sub == "" || contains(s, sub) }
	}

	type kv struct {
		name  string
		value float64
	}
	var out []kv

	for _, s := range payload.Stats {
		if !match(s.Name) {
			continue
		}
		// common request-related counters include:
		//   http.downstream_rq_total
		//   http.downstream_rq_2xx / 3xx / 4xx / 5xx
		//   cluster.<name>.upstream_rq_total
		//   cluster.<name>.upstream_rq_<code>xx
		//   listener.<name>.downstream_cx_total (connections)
		val := toFloat64(s.Value)
		out = append(out, kv{s.Name, val})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	for _, kv := range out {
		fmt.Printf("%-80s %12.0f\n", kv.name, kv.value)
	}
}

func toFloat64(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case json.Number:
		f, _ := t.Float64()
		return f
	case string:
		var n json.Number = json.Number(t)
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}

func contains(s, sub string) bool {
	// simple ASCII substring check; fast enough here
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}
