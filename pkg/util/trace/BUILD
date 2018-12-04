load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["traceutil.go"],
    importpath = "k8s.io/kubernetes/pkg/util/trace",
    visibility = ["//visibility:public"],
    deps = [
        "//staging/src/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/openzipkin/zipkin-go:go_default_library",
        "//vendor/github.com/openzipkin/zipkin-go/reporter/http:go_default_library",
        "//vendor/go.opencensus.io/exporter/zipkin:go_default_library",
        "//vendor/go.opencensus.io/trace:go_default_library",
        "//vendor/go.opencensus.io/trace/propagation:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
