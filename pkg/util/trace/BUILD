load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["traceutil.go"],
    importpath = "k8s.io/kubernetes/pkg/util/trace",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/core:go_default_library",
        "//staging/src/k8s.io/api/core/v1:go_default_library",
        "//vendor/contrib.go.opencensus.io/exporter/ocagent:go_default_library",
        "//vendor/contrib.go.opencensus.io/exporter/stackdriver:go_default_library",
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
