# Copyright 2025 The weekly-rundown Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# -- rules_go --
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "9d14ab75bb53c9a5b81cc1b1a3e01c5b93f6d8be6cf2832f85c765e69e5b8f7a",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.52.0/rules_go-v0.52.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.52.0/rules_go-v0.52.0.zip",
    ],
)

load("@io_bazel_rules_go//:go_deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

# -- gazelle --
http_archive(
    name = "bazel_gazelle",
    sha256 = "55afbf91c71c4c114b1c138e9e6e7c0e4c4e9b1c0d8e7f6a5b4c3d2e1f0a9b8",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.40.0/bazel-gazelle-v0.40.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.40.0/bazel-gazelle-v0.40.0.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

go_register_toolchains()
gazelle_dependencies()

# -- External Go dependencies --
go_repository(
    name = "com_github_gorilla_mux",
    importpath = "github.com/gorilla/mux",
    sum = "h1:8lGlTH3vqDu6opRKVn0dMvJGOM83mkFR1bKJdSBU1kI=",
    version = "v1.8.1",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    importpath = "github.com/mattn/go-isatty",
    sum = "h1:tGGpYGYiCTcrjRoX4A7JBnXp6MqaG3JXZFJJiJHMd2A=",
    version = "v0.0.20",
)

go_repository(
    name = "org_modernc_sqlite",
    importpath = "modernc.org/sqlite",
    sum = "h1:78SBHpmvrqaGPMoLBvnZBUVKDZ4sbMMS6el2oVyEJj0=",
    version = "v1.29.0",
)

# -- Transitive deps for modernc.org/sqlite --
go_repository(
    name = "org_modernc_libc",
    importpath = "modernc.org/libc",
    sum = "h1:QmURVLtVOSFhwtfqRzow8jxbzn9FkmD0bMhKAJZwL1gY=",
    version = "v1.41.0",
)

go_repository(
    name = "org_modernc_gc_v3",
    importpath = "modernc.org/gc/v3",
    sum = "h1:rKn4DdzCdHj5iCyR3Lf4+wHvAsGjKYJWPJUaR8bRPkA=",
    version = "v3.0.0-20240107210532-573471604cb6",
)

go_repository(
    name = "org_modernc_mathutil",
    importpath = "modernc.org/mathutil",
    sum = "h1:EG7Xgf08m2m94wMjjDze22b6uDoftl1it3VVuVIGUED=",
    version = "v1.6.0",
)

go_repository(
    name = "org_modernc_memory",
    importpath = "modernc.org/memory",
    sum = "h1:fh3fspfyXrCzov/Lew8I98fPHUqkqEee9kktEfcNP0s=",
    version = "v1.7.2",
)

go_repository(
    name = "org_modernc_strutil",
    importpath = "modernc.org/strutil",
    sum = "h1:agBi6dtTIMYgZsO4HyMTQAETUzRjLDISNQEAI/LdvnA=",
    version = "v1.2.0",
)

go_repository(
    name = "org_modernc_token",
    importpath = "modernc.org/token",
    sum = "h1:rIiE7kixVdmZcSfEqtoNqLE0GGhSqCLAOiA6ezXZfDY=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_dustin_go_humanize",
    importpath = "github.com/dustin/go-humanize",
    sum = "h1:VnRNEVXXEhe3hA+4cgHwdJyaFQBVV0hbDZViaJYnS3M=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_google_uuid",
    importpath = "github.com/google/uuid",
    sum = "h1:KDhWXfEtU9F29gXGdKMEgGtFBkGdVUNnFjWGKHQmQxs=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_hashicorp_golang_lru_v2",
    importpath = "github.com/hashicorp/golang-lru/v2",
    sum = "h1:4+BFUobTvtG+xAGPZhtLTUrjRaJHQpaJcfgFJPDKKJI=",
    version = "v2.0.7",
)

go_repository(
    name = "org_modernc_strftime",
    importpath = "github.com/ncruces/go-strftime",
    sum = "h1:giXQMjKvTJOcVtEMJNUCUcDJhoP3RaUfTbiZpOmpzYE=",
    version = "v0.1.9",
)

go_repository(
    name = "com_github_remymodpheng_bigfft",
    importpath = "github.com/remyoudompheng/bigfft",
    sum = "h1:W05bfYs+nYJPedvOJP4Bk5OSjQxHfSMp0KkJHdVfayg=",
    version = "v0.0.0-20230129092748-24d4a6f8daec",
)

go_repository(
    name = "org_golang_x_sys",
    importpath = "golang.org/x/sys",
    sum = "h1:DBtaqZIcRdRDrbanF9y1f3D7bBQZkh6baRfYCfGfEu8=",
    version = "v0.16.0",
)

workspace(name = "weekly_rundown")
