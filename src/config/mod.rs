use crate::utils::{kebab, read_file};

use anyhow::Result;
use chrono::{Local, NaiveDate, NaiveDateTime};
use serde::{Deserialize, Serialize};
use std::path::Path;

#[derive(Debug, Deserialize, Serialize)]
#[serde(default, rename_all = "kebab-case")]
pub struct Config {
    pub(crate) src: Option<toml::Value>,

    // Metadata
    pub title: String,
    pub authors: Vec<String>,
    pub date: Option<String>,

    // Subtables
    pub bib: BibConfig,
    pub build: BuildConfig,
    pub latex: LatexConfig,
    pub style: StyleConfig,
}

impl Config {
    pub fn from_file<P: AsRef<Path>>(path: P) -> Result<Self> {
        let config_content = read_file(path)?;
        Ok(toml::from_str(&config_content)?)
    }

    pub(crate) fn latex_header(&self) -> Option<String> {
        if self.latex.head.is_empty() && self.latex.packages.is_empty() {
            None
        } else {
            let mut header = String::new();
            header.push_str(&self.latex.head);
            let packages: String = self
                .latex
                .packages
                .iter()
                .map(|package| format!("\\usepackage{{{}}}\n", package))
                .collect();
            header.push_str(&packages);
            Some(header)
        }
    }

    pub(crate) fn date(&self) -> String {
        match self.date.as_deref() {
            Some("now") => {
                let now = Local::now();
                now.format(&self.style.date_format).to_string()
            }
            Some(date) => {
                // NOTE: When formatting with invalid format strings (e.g. time formatters when
                // time is not available), chrono panics with a really vague message. I am unable
                // to fix this myself and will just have to hope this does not happen.
                //
                // Related issue: https://github.com/chronotope/chrono/issues/575
                // Related PR: https://github.com/chronotope/chrono/pull/614
                if let Ok(dt) = date.parse::<NaiveDateTime>() {
                    return dt.format(&self.style.date_format).to_string();
                }
                if let Ok(dt) = date.parse::<NaiveDate>() {
                    return dt.format(&self.style.date_format).to_string();
                }
                date.to_owned()
            }
            None => String::default(),
        }
    }

    pub fn filename(&self) -> String {
        self.build
            .filename
            .to_owned()
            .unwrap_or_else(|| kebab(&self.title))
    }
}

impl Default for Config {
    fn default() -> Self {
        Self {
            src: None,
            title: "Document title".to_owned(),
            authors: vec![],
            date: None,
            bib: BibConfig::default(),
            build: BuildConfig::default(),
            latex: LatexConfig::default(),
            style: StyleConfig::default(),
        }
    }
}

#[derive(Debug, Deserialize, Serialize)]
#[serde(default, rename_all = "kebab-case")]
pub struct BibConfig {
    pub csl: String,
    pub src: Option<String>,
}

impl Default for BibConfig {
    fn default() -> Self {
        Self {
            csl: "apa".to_owned(),
            src: None,
        }
    }
}

#[derive(Debug, Default, Deserialize, Serialize)]
#[serde(default, rename_all = "kebab-case")]
pub struct BuildConfig {
    pub filename: Option<String>,
}

#[derive(Debug, Deserialize, Serialize)]
#[serde(default, rename_all = "kebab-case")]
pub struct StyleConfig {
    pub number_sections: bool,
    pub date_format: String,
    pub document_class: Option<String>,
}

impl Default for StyleConfig {
    fn default() -> Self {
        Self {
            number_sections: false,
            date_format: "%e %B %Y".to_owned(),
            document_class: None,
        }
    }
}

#[derive(Debug, Default, Deserialize, Serialize)]
#[serde(default, rename_all = "kebab-case")]
pub struct LatexConfig {
    pub head: String,
    pub packages: Vec<String>,
}
