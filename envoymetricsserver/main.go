package main

import (
	"flag"
	"log"
	"net"

	metrics_svc "github.com/envoyproxy/go-control-plane/envoy/service/metrics/v3"
	"google.golang.org/grpc"
)

type server struct {
}

func (s *server) StreamMetrics(stream metrics_svc.MetricsService_StreamMetricsServer) error {
	for {
		msg, err := stream.Recv()
		log.Printf("Got message from Envoy: %v", msg)
		if err != nil {
			return err
		}
		for _, set := range msg.EnvoyMetrics {
			for _, m := range set.Metric {
				// Extract counters that look like request stats
				if c := m.GetCounter(); c != nil {
					name := c.Exemplar.String()
					if looksLikeRq(name) {
						var val *float64
						if c.Value != nil {
							val = c.Value
						}
						log.Printf("[counter] %s = %d", name, val)
					}
				}
				// You can also inspect gauges & histograms:
				_ = m.GetGauge()
				_ = m.GetHistogram()
			}
		}
	}
}

func looksLikeRq(name string) bool {
	// Common patterns; adjust as needed
	return contains(name, "downstream_rq_") ||
		contains(name, "upstream_rq_") ||
		contains(name, ".rq_total") ||
		name == "http.downstream_rq_total"
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}

func main() {
	addr := flag.String("addr", "172.17.0.1:9999", "listen address")
	flag.Parse()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	grpcSrv := grpc.NewServer()
	metrics_svc.RegisterMetricsServiceServer(grpcSrv, &server{})
	log.Printf("MetricsService listening on %s", *addr)
	if err := grpcSrv.Serve(l); err != nil {
		log.Fatal(err)
	}
}
