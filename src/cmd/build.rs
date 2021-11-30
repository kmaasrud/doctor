use anyhow::{Context, Result};
use mdoc::{utils::write_file, DocumentBuilder};
use std::path::{Path, PathBuf};

pub fn build(path: Option<PathBuf>) -> Result<()> {
    let builder = DocumentBuilder::new();
    let doc = match path {
        Some(path) => builder.source(path).build()?,
        None => builder.build()?,
    };

    let pdf_data = doc.build()?;

    write_file(
        &Path::new(&doc.config.build.filename).with_extension("pdf"),
        &pdf_data,
    )
    .context("Could not write to PDF file")?;

    Ok(())
}
