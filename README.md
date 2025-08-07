# envtest

This crate provides a convenient Rust wrapper around the Go `envtest` package via the `rust2go` binding generator.  
It allows you to spin up a temporary Kubernetes control plane for integration testing and then clean it up cleanly.

## Features

- **Create** an isolated test environment.
- **Destroy** the environment automatically when the `Server` instance is dropped.
- **Retrieve** the kubeconfig as a strongly‑typed Rust structure.

## Usage

```rust
use envtest::Environment;
use kube::config::Kubeconfig;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. Build the default environment (uses latest envtest binaries)
    let env = Environment::default();

    // 2. Create the environment – this starts a temporary control plane
    let server = env.create()?;

    // 3. Consume the kubeconfig.
    let kubeconfig: Kubeconfig = server.kubeconfig()?;

    // 4. Use the kubeconfig to interact with Kubernetes
    let client = Client::try_from(kubeconfig)?;

    // 5. When `server` goes out of scope, the environment will be automatically destroyed.
    Ok(())
}
```

### Customizing the Environment

You can tweak the `Environment` configuration before calling `create`.  
`Environment` contains two nested structs – `CRDInstallOptions` and
`BinaryAssetsSettings` – that let you control which CRDs are installed
and how the envtest binaries are obtained.

* `crd_install_options.paths` – paths to directories or files containing CRDs. 
* `crd_install_options.crds` – list of CRD strings to install. 
* `crd_install_options.error_if_path_missing` – whether to fail if a path is not found.  
* `binary_assets_settings.download_binary_assets` – set to `false` to use binaries already available on the system (e.g., via `KUBEBUILDER_ASSETS`).  
* `binary_assets_settings.binary_assets_directory` – directory where the binaries will be stored or looked up.  
* `binary_assets_settings.download_binary_assets_version` – specific version to download.  
* `binary_assets_settings.download_binary_assets_index_url` – URL pointing to the envtest releases YAML.

#### Using `with_crds`

The `Environment::with_crds` method accepts a list of `CustomResourceDefinition`s to pre-install it with the environment. This works well with `kube` CRD generation trait and can be used to add custom CRDs without the need to store them in a separate file.

```rust
use envtest::{Environment, BinaryAssetsSettings};
use k8s_openapi::apiextensions_apiserver::pkg::apis::apiextensions::v1::CustomResourceDefinition;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Build the environment
    let env = Environment::default().with_crds(vec![CustomResourceDefinition::default()])?;

    // Create the test server with pre-defined CRDs
    let server = env.create()?;

    Ok(())
}
```

### Cleaning Up

The `Server` struct implements `Drop`, so the test environment is cleaned automatically when it is dropped.  
If you need explicit control:

```rust
let server = env.create()?;
server.destroy()?;
```

## Building the Bindings

The crate uses [`rust2go`][] to generate the Go bindings at build time. This crate relies on `bindgen` crate to generate the bindings from C headers, requiring `clang` dependencys to be installed. Refer to bindgen [requirements][] for more details.  
Make sure `GO111MODULE=on` and that Go is available on your PATH.  

[requirements]: https://rust-lang.github.io/rust-bindgen/requirements.html#requirements
[`rust2go`]: https://github.com/ihciah/rust2go

## License

MIT
