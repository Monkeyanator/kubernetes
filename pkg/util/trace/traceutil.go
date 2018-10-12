package traceutil

import (
	"context"
	"encoding/base64"
	"errors"

	"contrib.go.opencensus.io/exporter/stackdriver"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/core"
)

// SpanFromPodEncodedContext takes a Pod to extract trace context from and the desired Span name and
// constructs a new Span from this information
func SpanFromPodEncodedContext(pod *v1.Pod, name string) (ctx context.Context, result *trace.Span, err error) {

	// If there is no context encoded in the pod, error out
	if pod.TraceContext == "" {
		return context.Background(), &trace.Span{}, errors.New("could not extract trace context from given pod object")
	}

	decodedContextBytes, err := base64.StdEncoding.DecodeString(pod.TraceContext)
	if err != nil {
		return context.Background(), &trace.Span{}, err
	}

	remoteContext, ok := propagation.FromBinary(decodedContextBytes)
	if !ok {
		return context.Background(), &trace.Span{}, errors.New("could not convert raw bytes to trace")
	}

	newCtx, newSpan := trace.StartSpanWithRemoteParent(context.Background(), name, remoteContext)
	return newCtx, newSpan, nil
}

// EncodeSpanContextIntoPod takes a pointer to a pod and a trace context to embed
// Base64 encodes the wire format for the SpanContext, and puts it in the pod's TraceContext field
func EncodeSpanContextIntoPod(pod *core.Pod, spanContext trace.SpanContext) error {

	if string(pod.Name) == "" {
		pod.TraceContext = ""
		return errors.New("will not encode span into pod without name")
	}

	rawContextBytes := propagation.Binary(spanContext)
	encodedContext := base64.StdEncoding.EncodeToString(rawContextBytes)
	pod.TraceContext = encodedContext

	return nil
}

// DefaultExporter returns the default trace exporter for the project
// This is Stackdriver at the moment, but will be the OpenCensus agent
func DefaultExporter() (exporter trace.Exporter, err error) {
	// Create an register a OpenCensus
	// Stackdriver Trace exporter.
	exporter, err = stackdriver.NewExporter(stackdriver.Options{
		ProjectID: "samnaser-gke-dev-217421",
	})

	return exporter, err
}
