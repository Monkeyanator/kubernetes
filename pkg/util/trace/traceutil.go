package traceutil

import (
	"context"
	"encoding/base64"
	"errors"
	"log"

	"github.com/golang/glog"
	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/core"

	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

// SpanContextFromPodEncodedContext takes a pod to extract a SpanContext from and returns the decoded SpanContext
func SpanContextFromPodEncodedContext(pod *v1.Pod) (spanContext trace.SpanContext, err error) {

	glog.Errorf("span-context-from-pod-encoded")

	decodedContextBytes, err := base64.StdEncoding.DecodeString(pod.ObjectMeta.TraceContext)
	if err != nil {
		return trace.SpanContext{}, err
	}

	spanContext, ok := propagation.FromBinary(decodedContextBytes)
	if !ok {
		return trace.SpanContext{}, errors.New("could not convert raw bytes to trace")
	}

	return spanContext, nil

}

// SpanFromPodEncodedContext takes a Pod to extract trace context from and the desired Span name and
// constructs a new Span from this information
func SpanFromPodEncodedContext(pod *v1.Pod, name string) (ctx context.Context, result *trace.Span, err error) {

	glog.Errorf("span-from-pod-encoded-context")

	// If there is no context encoded in the pod, error out
	spanFromEncodedContext, err := SpanContextFromPodEncodedContext(pod)
	if err != nil {
		return context.Background(), &trace.Span{}, err
	}

	newCtx, newSpan := trace.StartSpanWithRemoteParent(context.Background(), name, spanFromEncodedContext)
	return newCtx, newSpan, nil
}

// EncodeSpanContextIntoPod takes a pointer to a pod and a trace context to embed
// Base64 encodes the wire format for the SpanContext, and puts it in the pod's TraceContext field
func EncodeSpanContextIntoPod(pod *core.Pod, spanContext trace.SpanContext) error {

	glog.Errorf("encoding span context into pod")

	rawContextBytes := propagation.Binary(spanContext)
	encodedContext := base64.StdEncoding.EncodeToString(rawContextBytes)
	pod.TraceContext = encodedContext

	return nil
}

// SpanContextToBase64String takes context and encodes it to a string
func SpanContextToBase64String(spanContext trace.SpanContext) string {

	rawContextBytes := propagation.Binary(spanContext)
	encodedContext := base64.StdEncoding.EncodeToString(rawContextBytes)

	return encodedContext
}

// InitializeExporter returns the default trace exporter for the project
// This is Stackdriver at the moment, but will be the OpenCensus agent
func InitializeExporter() error {

	glog.Errorf("default exporter created")

	// Stackdriver Trace exporter.
	// exporter, err := stackdriver.NewExporter(stackdriver.Options{})
	// if err != nil {
	// 	return err
	// }

	// Create the Zipkin exporter.
	localEndpoint, err := openzipkin.NewEndpoint("kubernetes-component", "192.168.1.5:5454")
	if err != nil {
		log.Fatalf("Failed to create the local zipkinEndpoint: %v", err)
	}
	reporter := zipkinHTTP.NewReporter("http://35.193.38.26:9411/api/v2/spans")
	ze := zipkin.NewExporter(reporter, localEndpoint)

	trace.RegisterExporter(ze)

	return nil
}

// SpanContextFromBase64String takes string and returns decoded context from it
func SpanContextFromBase64String(stringEncodedContext string) (spanContext trace.SpanContext, err error) {

	decodedContextBytes, err := base64.StdEncoding.DecodeString(stringEncodedContext)
	if err != nil {
		return trace.SpanContext{}, err
	}

	spanContext, ok := propagation.FromBinary(decodedContextBytes)
	if !ok {
		return trace.SpanContext{}, errors.New("could not convert raw bytes to trace")
	}

	return spanContext, nil

}
