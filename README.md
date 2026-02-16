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
  - [Explicit Cleanup](#explicit-cleanup)
- [Building the Bindings](#building-the-bindings)
- [License](#license)

---

## Features

- **Create** an isolated test environment with a fully‑working control plane.
- **Destroy** the environment automatically when the `Server` instance is dropped.
- **Retrieve** the kubeconfig as a strongly‑typed `kube::config::Kubeconfig`.
- **Pre‑install** provide user CRDs, either from files or in‑memory definitions.

---

## Getting Started

### Installation

Add `envtest` to your Cargo.toml:

```toml
[dependencies]
envtest = "0.1"
kube = { version = "1" }
```

> **Note**: `rust2go` requires a working Go toolchain and `clang` for the bindgen step.  
> Make sure that Go (`GO111MODULE=on`) is available on your PATH.

### Basic Usage

```rust
use envtest::Environment;
use kube::{Client, config::Kubeconfig};

// Implement extension to define Kubeconfig and Client returned by the server across your crate and kube version
envtest::impl_client_ext!(Kubeconfig, Client);

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. Build the default environment (latest envtest binaries)
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

`Environment` exposes two nested structs:

| Field | Purpose |
|-------|---------|
| `crd_install_options.paths` | Paths to directories or files with CRDs. |
| `crd_install_options.crds` | In‑memory CRDs to install. |
| `crd_install_options.error_if_path_missing` | Fail if a a CRD path is missing. |
| `binary_assets_settings.download_binary_assets` | `false` → use binaries from `$KUBEBUILDER_ASSETS`. |
| `binary_assets_settings.binary_assets_directory` | Cache directory for downloaded binaries. |
| `binary_assets_settings.download_binary_assets_version` | Specific `envtest` version to download. |
| `binary_assets_settings.download_binary_assets_index_url` | URL pointing to the envtest release index. |

#### Using `with_crds`

```rust
use envtest::Environment;
use k8s_openapi::apiextensions_apiserver::pkg::apis::apiextensions::v1::CustomResourceDefinition;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let env = Environment::default()
        .with_crds(vec![CustomResourceDefinition::default()])?;

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

### Typed Client Helper

To avoid manually specifying `Kubeconfig` and `Client` instance for each setup, invoke `impl_client_ext!` once for your version of `kube`:

```rust
use envtest::Environment;
use kube::{Client, config::Kubeconfig};

envtest::impl_client_ext!(Kubeconfig, Client);

#[tokio::test]
async fn build_client_in_test() {
    let server = Environment::default().create().unwrap();
    let _client = server.client().unwrap(); // Creates Client instance directly
    server.destroy().unwrap();
}
```

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
