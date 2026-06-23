use std::{env, fs, path::PathBuf};

use anyhow::{Context, Result};
use serde::Deserialize;

use crate::theme::Theme;

#[derive(Debug, Clone, Default, Deserialize)]
pub struct Config {
    pub theme: Option<Theme>,
}

impl Config {
    pub fn load(explicit_path: Option<PathBuf>) -> Result<Self> {
        let Some(path) = explicit_path.or_else(default_config_path) else {
            return Ok(Self::default());
        };

        if !path.exists() {
            return Ok(Self::default());
        }

        let raw = fs::read_to_string(&path)
            .with_context(|| format!("failed to read config file {}", path.display()))?;
        toml::from_str(&raw)
            .with_context(|| format!("failed to parse config file {}", path.display()))
    }

    pub fn theme_or_default(self) -> Theme {
        self.theme.unwrap_or_default()
    }
}

fn default_config_path() -> Option<PathBuf> {
    let home = env::var_os("HOME")?;
    Some(
        PathBuf::from(home)
            .join(".config")
            .join("mdview")
            .join("config.toml"),
    )
}
