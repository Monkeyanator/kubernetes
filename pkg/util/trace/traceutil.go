// Package traceutil provides various definitions and utilities that allow for
// common operations with our trace tooling, such as span creation, encoding, decoding,
// and enumeration of possible services.
package traceutil

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang/glog"
	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

// trace exporter configuration
const (
	DefaultTraceAddress     = "192.168.1.5"
	DefaultTracePort        = "5454"
	DefaultCollectorAddress = "http://35.193.38.26:9411/api/v2/spans"
)

// services a given span could export from
const (
	ServiceAPIServer         = "api-server"
	ServiceScheduler         = "scheduler"
	ServiceKubelet           = "kubelet"
	ServiceContainerdRuntime = "containerd-runtime"
	ServiceCRI               = "containerd-cri"
)

// ServiceType represents a logical service within Kubernetes
type ServiceType string

// InitializeExporter takes a ServiceType and sets the global OpenCensus exporter
// to export to that service on a specified Zipkin instance
func InitializeExporter(service ServiceType) error {

	glog.Infof("OpenCensus trace exporter initializing with service %s", string(service))

	// create zipkin exporter
	localEndpoint, err := openzipkin.NewEndpoint(string(service), fmt.Sprintf("%s:%s", DefaultTraceAddress, DefaultTracePort))
	if err != nil {
		glog.Errorf("failed to create the local zipkinEndpoint: %v", err)
	}
	reporter := zipkinHTTP.NewReporter(DefaultCollectorAddress)
	ze := zipkin.NewExporter(reporter, localEndpoint)

	trace.RegisterExporter(ze)

	return nil
}

// SpanFromEncodedContext takes an object to extract trace context from and the desired Span name and
// constructs a new Span from this information
func SpanFromEncodedContext(tracedResource meta.Object, name string) (ctx context.Context, result *trace.Span, err error) {

	glog.Infof("creating span from SpanContext encoded in object %s", tracedResource.GetName())
	spanFromEncodedContext, err := SpanContextFromEncodedContext(tracedResource)
	if err != nil {
		return context.Background(), &trace.Span{}, err
	}

	newCtx, newSpan := trace.StartSpanWithRemoteParent(context.Background(), name, spanFromEncodedContext)
	return newCtx, newSpan, nil
}

// SpanContextFromEncodedContext takes an object to extract an encoded SpanContext from and returns the decoded SpanContext
func SpanContextFromEncodedContext(tracedResource meta.Object) (spanContext trace.SpanContext, err error) {

	decodedContextBytes, err := base64.StdEncoding.DecodeString(tracedResource.GetTraceContext())
	if err != nil {
		return trace.SpanContext{}, err
	}

	spanContext, ok := propagation.FromBinary(decodedContextBytes)
	if !ok {
		return trace.SpanContext{}, fmt.Errorf("could not convert raw bytes to trace from object %s", tracedResource.GetName())
	}

	return spanContext, nil

}

// EncodeSpanContextIntoObject takes a pointer to an object and a trace context to embed
// Base64 encodes the wire format for the SpanContext, and puts it in the object's TraceContext field
func EncodeSpanContextIntoObject(tracedResource meta.Object, spanContext trace.SpanContext) error {

	glog.Infof("encoding serialized SpanContext into object %s", tracedResource.GetName())

	rawContextBytes := propagation.Binary(spanContext)
	encodedContext := base64.StdEncoding.EncodeToString(rawContextBytes)
	tracedResource.SetTraceContext(encodedContext)

	return nil
}

// EndRootObjectTraceWithName takes a traced resource, the final ServiceType, and the desired name
// and exports the corresponding root span into the specified tracing backend
func EndRootObjectTraceWithName(tracedResource meta.Object, service ServiceType, spanName string) {

	rootSpanContext, _ := SpanContextFromEncodedContext(tracedResource)
	spanData := &trace.SpanData{
		SpanContext:  rootSpanContext,
		ParentSpanID: trace.SpanID{0x0},
		Name:         spanName,
		StartTime:    tracedResource.GetCreationTimestamp().Time,
		EndTime:      time.Now(),
		Status:       trace.Status{Code: trace.StatusCodeOK},
	}

	// Must create a separate Zipkin exporter here since it's not possible to access the global exporter directly
	localEndpoint, err := openzipkin.NewEndpoint(string(service), fmt.Sprintf("%s:%s", DefaultTraceAddress, DefaultTracePort))
	if err != nil {
		log.Fatalf("Failed to create the local zipkinEndpoint: %v", err)
	}
	reporter := zipkinHTTP.NewReporter(DefaultCollectorAddress)
	ze := zipkin.NewExporter(reporter, localEndpoint)
	ze.ExportSpan(spanData)

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
