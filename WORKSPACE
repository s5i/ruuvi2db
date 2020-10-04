load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "b725e6497741d7fc2d55fcc29a276627d10e43fa5d0bb692692890ae30d98d00",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.24.3/rules_go-v0.24.3.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.24.3/rules_go-v0.24.3.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "b85f48fa105c4403326e9525ad2b2cc437babaa6e15a3fc0b1dbab0ab064bc7c",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.22.2/bazel-gazelle-v0.22.2.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.22.2/bazel-gazelle-v0.22.2.tar.gz",
    ],
)

http_archive(
    name = "com_google_protobuf",
    sha256 = "9748c0d90e54ea09e5e75fb7fac16edce15d2028d4356f32211cfa3c0e956564",
    strip_prefix = "protobuf-3.11.4",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.11.4.zip"],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@io_bazel_rules_go//extras:embed_data_deps.bzl", "go_embed_data_dependencies")

go_embed_data_dependencies()

go_rules_dependencies()

go_register_toolchains()

gazelle_dependencies()

protobuf_deps()

go_repository(
    name = "com_github_go_ble_ble",
    commit = "067514cd6e24acd5069d01b3f0712d58e666a4f5",
    importpath = "github.com/go-ble/ble",
)

go_repository(
    name = "com_github_influxdata_influxdb1_client",
    commit = "b269163b24ab8e62026d13a92aa988a7389c3b4e",
    importpath = "github.com/influxdata/influxdb1-client",
)

go_repository(
    name = "com_github_boltdb_bolt",
    commit = "fd01fc79c553a8e99d512a07e8e0c63d4a3ccfc5",
    importpath = "github.com/boltdb/bolt",
)

go_repository(
    name = "com_github_pkg_errors",
    commit = "614d223910a179a466c1767a985424175c39b465",
    importpath = "github.com/pkg/errors",
)

go_repository(
    name = "com_github_mgutz_logxi",
    commit = "aebf8a7d67ab4625e0fd4a665766fef9a709161b",
    importpath = "github.com/mgutz/logxi",
)

go_repository(
    name = "com_github_mgutz_ansi",
    commit = "d51e80ef957dba7f19388ca64afefbd5a096af30",
    importpath = "github.com/mgutz/ansi",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    commit = "cb30d6282491c185f77d9bec5d25de1bb61a06bc",
    importpath = "github.com/mattn/go-isatty",
)

go_repository(
    name = "com_github_mattn_go_colorable",
    commit = "f6c00982823144337e56cdb71c712eaac151d29c",
    importpath = "github.com/mattn/go-colorable",
)

