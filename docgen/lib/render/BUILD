package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = glob(["*.go"]),
    deps = [
        "//docgen/lib/proto:go_default_library",
        "//docgen/lib/render/templates:go_default_library",
        "@com_github_ghodss_yaml//:go_default_library",
        "@com_github_golang_protobuf//jsonpb:go_default_library",
        "@in_gopkg_yaml_v2//:go_default_library",
    ],
)
