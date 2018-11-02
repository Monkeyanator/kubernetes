//mple trace_quickstart creates traces incoming and outgoing requests.
package main

import (
	"context"
	"encoding/base64"
	"log"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

func main() {

	log.Println("Execution begin...")
	traceContext := os.Getenv("KUBERNETES_TRACE_CONTEXT")
	log.Println("Downward API passed trace context: ", traceContext)
	log.Println("Another test")

	// Create an register a OpenCensus
	// Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: "samnaser-gke-dev-217421",
	})
	if err != nil {
		log.Fatal(err)
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(exporter)

	log.Println("Stackdriver exporter created.")

	decodedContextBytes, err := base64.StdEncoding.DecodeString(traceContext)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Decoded context.")

	spanContext, ok := propagation.FromBinary(decodedContextBytes)
	if !ok {
		log.Fatalf("could not convert raw bytes to trace")
	}

	log.Println("Trace ID: ", spanContext.TraceID)
	_, span := trace.StartSpanWithRemoteParent(context.Background(), "Deep roots are not reached by the frost", spanContext)
	time.Sleep(2 * time.Second)
	span.End()

	log.Println("Span ended.")
	time.Sleep(time.Minute * 2)

}
