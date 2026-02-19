# envtest

A lightweight, type‑safe wrapper around the Kubernetes `envtest` Go package that lets you spin up a temporary control plane from Rust.

`envtest` aims to make integration testing with real Kubernetes components straightforward. This project is based on the `envtest` Go package, but provides a Rust interface for the library usage.

---

## Table of Contents

- [Features](#features)
- [Getting Started](#getting-started)
  - [Installation](#installation)
  - [Basic Usage](#basic-usage)
  - [Customizing the Environment](#customizing-the-environment)
    - [Using `with_crds`](#using-with_crds)
  - [Explicit Cleanup](#explicit-cleanup)
  - [Usage in testing](#usage-in-testing)
- [Building the Bindings](#building-the-bindings)
- [License](#license)

---

## Features

- **Create** an isolated test environment with a fully‑working control plane.
- **Destroy** the environment automatically when the `Server` instance is dropped.
- **Retrieve** the kubeconfig as a strongly‑typed `kube::config::Kubeconfig`.
- **Pre‑install** user CRDs from in‑memory definitions.

---

## Getting Started

### Installation

Add `envtest` to your Cargo.toml:

```toml
[dependencies]
envtest = "0.1"
```

> **Note**: `rust2go` requires a working Go toolchain and `clang` for the bindgen step.  
> Make sure that Go (`GO111MODULE=on`) is available on your PATH.

### Basic Usage

```rust
use envtest::Environment;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. Build the default environment
    let env = Environment::default();

    // 2. Spin up a temporary control plane
    let server = env.create()?;

    // 3. Retrieve a strongly‑typed kubeconfig
    let kubeconfig = server.kubeconfig()?;

    // 4. Create a `kube` client
    let client = server.client()?;

    // 5. `server` is dropped at the end of scope, cleaning up the control plane
    Ok(())
}
```

### Customizing the Environment

You can tweak binary download behavior through `binary_assets_settings`:

| Field | Purpose |
|-------|---------|
| `crd_install_options.paths` | Paths to directories or files with CRDs. |
| `crd_install_options.crds` | In‑memory CRDs to install. |
| `crd_install_options.error_if_path_missing` | Fail if a a CRD path is missing. |
| `binary_assets_settings.download_binary_assets` | `false` → use binaries from `$KUBEBUILDER_ASSETS`. |
| `binary_assets_settings.binary_assets_directory` | Cache directory for downloaded binaries. |
| `binary_assets_settings.download_binary_assets_version` | Specific `envtest` version to download. |
| `binary_assets_settings.download_binary_assets_index_url` | URL pointing to the envtest release index. |

```rust
use envtest::Environment;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut env = Environment::default();
    env.binary_assets_settings.download_binary_assets_version = Some("1.32.0".to_string());
    env.binary_assets_settings.binary_assets_directory = Some(".cache/envtest".to_string());
    let server = env.create()?;
    Ok(())
}
```

#### Using `with_crds`

```rust
use envtest::Environment;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let crd = MyCustomResource::crd();

    let env = Environment::default()
        .with_crds(crd)?;

    let server = env.create()?;
    Ok(())
}
```

### Explicit Cleanup

The `Server` implements `Drop`, but you can destroy it manually:

```rust
let server = env.create()?;
server.destroy()?;
```


### Usage in testing

`envtest` is most useful in combination with the `kube` feature (enabled by default), which provides a strongly typed API client for interacting with the Kubernetes API.

```rust
use envtest::Environment;
use kube::{Api, api::{DeleteParams, PostParams}};
use k8s_openapi::api::core::v1::Namespace;

#[tokio::test]
async fn work_with_namespace_with_kube_client() -> Result<(), Box<dyn std::error::Error>> {
    let server = Environment::default().create()?;
    let namespaces: Api<Namespace> = Api::all(server.client()?);
    namespaces
        .create(
            &PostParams::default(),
            &Namespace {
                metadata: kube::core::ObjectMeta {
                    name: Some("envtest-e2e".to_string()),
                    ..Default::default()
                },
                ..Default::default()
            },
        )
        .await?;

    namespaces
        .delete("envtest-e2e", &DeleteParams::default())
        .await?;

    Ok(())
}
```

For test suites, prefer one `Environment` per test to avoid shared state between cases.

---

## Building the Bindings

`envtest` generates Go bindings at compile time via `rust2go`. The build process relies on:

- [`rust2go`][] (which uses `bindgen` under the hood)
- A working Go toolchain (`GO111MODULE=on`)
- `clang` (for bindgen)

Refer to the [bindgen requirements](https://rust-lang.github.io/rust-bindgen/requirements.html#requirements) for more details.

[`rust2go`]: https://github.com/ihciah/rust2go

---

## License

MIT license – see the [LICENSE](LICENSE) file for details.

---
