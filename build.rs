use std::fs;
use std::path::Path;
use std::time::SystemTime;

macro_rules! watch_files {
    ($($p:literal),* $(,)?) => {
        $(println!("cargo:rerun-if-changed={}", $p);)*
    };
}

fn modified(path: &Path) -> Option<SystemTime> {
    fs::metadata(path).ok()?.modified().ok()
}

fn should_regen(src: &Path, dst: &Path) -> Option<bool> {
    Some(modified(src)? > modified(dst)?)
}

fn main() {
    // Ensure build script reruns when Rust API definitions change.
    watch_files!("src", "go/go.sum", "go/go.mod", "go/impl.go",);

    let mut builder = rust2go::Builder::new().with_go_src("go");

    let src = Path::new("./src/lib.rs");
    let dst = Path::new("./go/gen.go");
    if let Some(true) | None = should_regen(src, dst) {
        builder = builder.with_regen_arg(rust2go::RegenArgs {
            src: src.display().to_string(),
            dst: dst.display().to_string(),
            ..Default::default()
        });
    }

    builder.build();
}
