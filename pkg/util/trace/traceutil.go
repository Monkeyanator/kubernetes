// Package traceutil provides various definitions and utilities that allow for
// common operations with our trace tooling, such as span creation, encoding, decoding,
// and enumeration of possible services.
package traceutil

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/golang/glog"
	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/core"

	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

// potential services a given span could export from
const (
	ServiceAPIServer         = "api-server"
	ServiceScheduler         = "scheduler"
	ServiceKubelet           = "kubelet"
	ServiceContainerdRuntime = "containerd-runtime"
	ServiceCRI               = "containerd-cri"
)

// ServiceType represents a logical service within Kubernetes
type ServiceType string

// TracedResource represents an object that can have serialized trace context inserted and embedded
type TracedResource interface {
	retrieveEmbeddedContext() string
	embedContext(string)
}

// InitializeExporter takes a ServiceType and sets the global OpenCensus exporter
// to export to that service on a specified Zipkin instance
func InitializeExporter(service ServiceType) error {

	glog.Infof("OpenCensus trace exporter initializing with service %s", string(service))

	// create zipkin exporter
	localEndpoint, err := openzipkin.NewEndpoint(string(service), "192.168.1.5:5454")
	if err != nil {
		glog.Errorf("failed to create the local zipkinEndpoint: %v", err)
	}
	reporter := zipkinHTTP.NewReporter("http://35.193.38.26:9411/api/v2/spans")
	ze := zipkin.NewExporter(reporter, localEndpoint)

	trace.RegisterExporter(ze)

	return nil
}

// SpanFromPodEncodedContext takes a Pod to extract trace context from and the desired Span name and
// constructs a new Span from this information
func SpanFromPodEncodedContext(pod *v1.Pod, name string) (ctx context.Context, result *trace.Span, err error) {

	glog.Infof("creating span from SpanContext encoded in pod %s", pod.Name)
	pod.SetResourceVersion("asdf")
	spanFromEncodedContext, err := SpanContextFromPodEncodedContext(pod)
	if err != nil {
		return context.Background(), &trace.Span{}, err
	}

	newCtx, newSpan := trace.StartSpanWithRemoteParent(context.Background(), name, spanFromEncodedContext)
	return newCtx, newSpan, nil
}

// SpanContextFromPodEncodedContext takes a pod to extract an encoded SpanContext from and returns the decoded SpanContext
func SpanContextFromPodEncodedContext(pod *v1.Pod) (spanContext trace.SpanContext, err error) {

	decodedContextBytes, err := base64.StdEncoding.DecodeString(pod.ObjectMeta.TraceContext)
	if err != nil {
		return trace.SpanContext{}, err
	}

	spanContext, ok := propagation.FromBinary(decodedContextBytes)
	if !ok {
		return trace.SpanContext{}, fmt.Errorf("could not convert raw bytes to trace from pod %s", pod.Name)
	}

	return spanContext, nil

}

// EncodeSpanContextIntoPod takes a pointer to a pod and a trace context to embed
// Base64 encodes the wire format for the SpanContext, and puts it in the pod's TraceContext field
func EncodeSpanContextIntoPod(pod *core.Pod, spanContext trace.SpanContext) error {

	glog.Infof("encoding serialized SpanContext into pod %s", pod.Name)

	rawContextBytes := propagation.Binary(spanContext)
	encodedContext := base64.StdEncoding.EncodeToString(rawContextBytes)
	pod.TraceContext = encodedContext

	return nil
}

// SpanContextToBase64String takes a SpanContext and returns a serialized string
func SpanContextToBase64String(spanContext trace.SpanContext) string {

	rawContextBytes := propagation.Binary(spanContext)
	encodedContext := base64.StdEncoding.EncodeToString(rawContextBytes)

	return encodedContext
}

// SpanContextFromBase64String takes string and returns decoded SpanContext
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
