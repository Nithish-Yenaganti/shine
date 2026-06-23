use std::{
    fs,
    io::{self, IsTerminal, Read},
    path::{Path, PathBuf},
    time::SystemTime,
};

use anyhow::{anyhow, Context, Result};
use clap::{CommandFactory, Parser, ValueEnum};
use clap_complete::{generate, Shell};

use crate::{
    config::Config,
    model::Document,
    parser::parse_document,
    render::{plain_text, render_document_with_theme, RenderedDocument},
    theme::Theme,
    tui,
};

#[derive(Debug, Parser)]
#[command(name = "mdview")]
#[command(about = "Preview Markdown beautifully inside the terminal.")]
pub struct Args {
    /// Markdown file to preview. If omitted, mdview reads from stdin.
    pub path: Option<PathBuf>,

    /// Re-render the file when it changes.
    #[arg(short, long)]
    pub watch: bool,

    /// Print the rendered document and exit instead of launching the TUI.
    #[arg(long)]
    pub print: bool,

    /// Path to a TOML config file.
    #[arg(long)]
    pub config: Option<PathBuf>,

    /// Built-in theme preset. Config file theme values override this.
    #[arg(long, value_enum, default_value_t = ThemePreset::Midnight)]
    pub theme: ThemePreset,

    /// Disable alternate screen mode.
    #[arg(long)]
    pub no_alt_screen: bool,

    /// Render width for --print mode.
    #[arg(long, default_value_t = 88)]
    pub width: usize,

    #[command(subcommand)]
    pub command: Option<Command>,
}

#[derive(Debug, Clone, Copy, ValueEnum)]
pub enum ThemePreset {
    Midnight,
    Daylight,
    Mono,
}

#[derive(Debug, Clone, clap::Subcommand)]
pub enum Command {
    /// Generate shell completions.
    Completions { shell: Shell },
}

#[derive(Debug, Clone)]
pub enum Source {
    File(PathBuf),
    Stdin { name: String, contents: String },
}

#[derive(Debug, Clone)]
pub struct LoadedDocument {
    pub document: Document,
    pub modified: Option<SystemTime>,
}

pub fn run() -> Result<()> {
    let args = Args::parse();
    if let Some(Command::Completions { shell }) = args.command {
        let mut command = Args::command();
        generate(shell, &mut command, "mdview", &mut io::stdout());
        return Ok(());
    }

    let config = Config::load(args.config.clone())?;
    let theme = config.theme.unwrap_or_else(|| preset_theme(args.theme));
    let source = resolve_source(args.path.clone())?;

    if args.print {
        let loaded = load_document(&source)?;
        let rendered = render_document_with_theme(&loaded.document, args.width, &theme);
        print!("{}", plain_text(&rendered));
        return Ok(());
    }

    tui::run(tui::PreviewApp::new(
        source,
        theme,
        args.watch,
        !args.no_alt_screen,
    )?)
}

fn preset_theme(theme: ThemePreset) -> crate::theme::Theme {
    match theme {
        ThemePreset::Midnight => crate::theme::Theme::default(),
        ThemePreset::Daylight => crate::theme::Theme::light(),
        ThemePreset::Mono => crate::theme::Theme::mono(),
    }
}

pub fn resolve_source(path: Option<PathBuf>) -> Result<Source> {
    if let Some(path) = path {
        return Ok(Source::File(path));
    }

    if io::stdin().is_terminal() {
        return Err(anyhow!(
            "provide a Markdown file or pipe Markdown into stdin"
        ));
    }

    let mut contents = String::new();
    io::stdin()
        .read_to_string(&mut contents)
        .context("failed to read Markdown from stdin")?;
    Ok(Source::Stdin {
        name: "stdin".to_string(),
        contents,
    })
}

pub fn load_document(source: &Source) -> Result<LoadedDocument> {
    match source {
        Source::File(path) => {
            let raw = fs::read_to_string(path)
                .with_context(|| format!("failed to read Markdown file {}", path.display()))?;
            let modified = file_modified(path).ok();
            Ok(LoadedDocument {
                document: parse_document(&raw, Some(path.display().to_string())),
                modified,
            })
        }
        Source::Stdin { name, contents } => Ok(LoadedDocument {
            document: parse_document(contents, Some(name.clone())),
            modified: None,
        }),
    }
}

pub fn source_name(source: &Source) -> String {
    match source {
        Source::File(path) => path.display().to_string(),
        Source::Stdin { name, .. } => name.clone(),
    }
}

pub fn file_modified(path: &Path) -> Result<SystemTime> {
    Ok(fs::metadata(path)
        .with_context(|| format!("failed to inspect {}", path.display()))?
        .modified()
        .with_context(|| format!("failed to read modified time for {}", path.display()))?)
}

pub fn render_loaded(document: &Document, width: usize, theme: &Theme) -> RenderedDocument {
    render_document_with_theme(document, width, theme)
}
