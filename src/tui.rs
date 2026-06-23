use std::{
    io::{self, Stdout},
    path::Path,
    sync::mpsc::{self, Receiver},
    time::{Duration, SystemTime},
};

use anyhow::{Context, Result};
use crossterm::{
    event::{self, Event, KeyCode, KeyEvent, KeyModifiers},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use notify::{Config, RecommendedWatcher, RecursiveMode, Watcher};
use ratatui::{
    backend::CrosstermBackend,
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block as TuiBlock, Borders, Clear, List, ListItem as TuiListItem, Paragraph, Wrap},
    Frame, Terminal,
};

use crate::{
    app::{load_document, render_loaded, source_name, LoadedDocument, Source},
    model::Document,
    render::{line_to_tui, RenderedDocument},
    theme::Theme,
};

type Term = Terminal<CrosstermBackend<Stdout>>;

pub struct PreviewApp {
    source: Source,
    theme: Theme,
    watch: bool,
    use_alt_screen: bool,
    document: Document,
    rendered: RenderedDocument,
    last_modified: Option<SystemTime>,
    render_width: usize,
    scroll: usize,
    search_mode: bool,
    search_query: String,
    matches: Vec<usize>,
    current_match: usize,
    outline_open: bool,
    outline_index: usize,
    status: String,
    error: Option<String>,
    should_quit: bool,
}

impl PreviewApp {
    pub fn new(source: Source, theme: Theme, watch: bool, use_alt_screen: bool) -> Result<Self> {
        let loaded = load_document(&source)?;
        let rendered = render_loaded(&loaded.document, 88, &theme);
        let status = format!("loaded {}", source_name(&source));

        Ok(Self {
            source,
            theme,
            watch,
            use_alt_screen,
            document: loaded.document,
            rendered,
            last_modified: loaded.modified,
            render_width: 88,
            scroll: 0,
            search_mode: false,
            search_query: String::new(),
            matches: Vec::new(),
            current_match: 0,
            outline_open: false,
            outline_index: 0,
            status,
            error: None,
            should_quit: false,
        })
    }

    fn reload(&mut self) {
        match load_document(&self.source) {
            Ok(LoadedDocument { document, modified }) => {
                self.document = document;
                self.last_modified = modified;
                self.rerender(self.render_width);
                self.status = format!("reloaded {}", source_name(&self.source));
                self.error = None;
            }
            Err(error) => {
                self.error = Some(error.to_string());
            }
        }
    }

    fn rerender(&mut self, width: usize) {
        let width = width.max(24);
        if self.render_width != width || self.rendered.lines.is_empty() {
            self.render_width = width;
            self.rendered = render_loaded(&self.document, width, &self.theme);
            self.rebuild_matches();
        }
        self.clamp_scroll();
    }

    fn clamp_scroll(&mut self) {
        if self.rendered.lines.is_empty() {
            self.scroll = 0;
            return;
        }
        self.scroll = self.scroll.min(self.rendered.lines.len().saturating_sub(1));
    }

    fn handle_key(&mut self, key: KeyEvent, visible_height: usize) {
        if self.search_mode {
            self.handle_search_key(key);
            return;
        }

        if self.outline_open {
            self.handle_outline_key(key);
            return;
        }

        match key.code {
            KeyCode::Char('q') => self.should_quit = true,
            KeyCode::Char('r') => self.reload(),
            KeyCode::Char('/') => {
                self.search_mode = true;
                self.search_query.clear();
                self.rebuild_matches();
            }
            KeyCode::Char('n') => self.next_match(),
            KeyCode::Char('N') => self.previous_match(),
            KeyCode::Char('o') => {
                self.outline_open = true;
                self.outline_index = 0;
            }
            KeyCode::Char('g') => self.scroll = 0,
            KeyCode::Char('G') => {
                self.scroll = self.rendered.lines.len().saturating_sub(visible_height);
            }
            KeyCode::Char('j') | KeyCode::Down => self.scroll = self.scroll.saturating_add(1),
            KeyCode::Char('k') | KeyCode::Up => self.scroll = self.scroll.saturating_sub(1),
            KeyCode::PageDown => self.scroll = self.scroll.saturating_add(visible_height),
            KeyCode::PageUp => self.scroll = self.scroll.saturating_sub(visible_height),
            KeyCode::Home => self.scroll = 0,
            KeyCode::End => {
                self.scroll = self.rendered.lines.len().saturating_sub(visible_height);
            }
            KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => {
                self.should_quit = true;
            }
            _ => {}
        }
        self.clamp_scroll();
    }

    fn handle_search_key(&mut self, key: KeyEvent) {
        match key.code {
            KeyCode::Esc => {
                self.search_mode = false;
                self.search_query.clear();
                self.rebuild_matches();
            }
            KeyCode::Enter => {
                self.search_mode = false;
                self.jump_to_current_match();
            }
            KeyCode::Backspace => {
                self.search_query.pop();
                self.rebuild_matches();
                self.jump_to_current_match();
            }
            KeyCode::Char(ch) => {
                self.search_query.push(ch);
                self.rebuild_matches();
                self.jump_to_current_match();
            }
            _ => {}
        }
    }

    fn handle_outline_key(&mut self, key: KeyEvent) {
        match key.code {
            KeyCode::Esc | KeyCode::Char('o') => self.outline_open = false,
            KeyCode::Down | KeyCode::Char('j') => {
                self.outline_index = self
                    .outline_index
                    .saturating_add(1)
                    .min(self.document.headings.len().saturating_sub(1));
            }
            KeyCode::Up | KeyCode::Char('k') => {
                self.outline_index = self.outline_index.saturating_sub(1);
            }
            KeyCode::Enter => {
                if let Some(heading) = self.document.headings.get(self.outline_index) {
                    if let Some(line) = self.rendered.block_line_offsets.get(heading.block_index) {
                        self.scroll = *line;
                    }
                }
                self.outline_open = false;
            }
            _ => {}
        }
    }

    fn rebuild_matches(&mut self) {
        self.matches.clear();
        let query = self.search_query.trim().to_ascii_lowercase();
        if query.is_empty() {
            self.current_match = 0;
            return;
        }

        self.matches = self
            .rendered
            .lines
            .iter()
            .enumerate()
            .filter_map(|(index, line)| {
                line.text
                    .to_ascii_lowercase()
                    .contains(&query)
                    .then_some(index)
            })
            .collect();
        self.current_match = self.current_match.min(self.matches.len().saturating_sub(1));
    }

    fn next_match(&mut self) {
        if self.matches.is_empty() {
            return;
        }
        self.current_match = (self.current_match + 1) % self.matches.len();
        self.jump_to_current_match();
    }

    fn previous_match(&mut self) {
        if self.matches.is_empty() {
            return;
        }
        self.current_match = if self.current_match == 0 {
            self.matches.len() - 1
        } else {
            self.current_match - 1
        };
        self.jump_to_current_match();
    }

    fn jump_to_current_match(&mut self) {
        if let Some(line) = self.matches.get(self.current_match) {
            self.scroll = *line;
        }
    }

    fn status_text(&self) -> String {
        if let Some(error) = &self.error {
            return format!("error: {error}");
        }
        if self.search_mode {
            return format!("/{}  {} matches", self.search_query, self.matches.len());
        }

        let watch = if self.watch { "watch:on" } else { "watch:off" };
        let position = if self.rendered.lines.is_empty() {
            "0/0".to_string()
        } else {
            format!("{}/{}", self.scroll + 1, self.rendered.lines.len())
        };
        format!(
            "{} | {} blocks | {} headings | {} | {} | q quit / search o outline r reload",
            source_name(&self.source),
            self.document.block_count(),
            self.document.headings.len(),
            position,
            watch,
        )
    }
}

pub fn run(mut app: PreviewApp) -> Result<()> {
    enable_raw_mode().context("failed to enable terminal raw mode")?;
    let mut stdout = io::stdout();
    if app.use_alt_screen {
        execute!(stdout, EnterAlternateScreen).context("failed to enter alternate screen")?;
    }
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend).context("failed to create terminal backend")?;

    let result = run_loop(&mut terminal, &mut app);

    disable_raw_mode().ok();
    if app.use_alt_screen {
        execute!(terminal.backend_mut(), LeaveAlternateScreen).ok();
    }
    terminal.show_cursor().ok();

    result
}

fn run_loop(terminal: &mut Term, app: &mut PreviewApp) -> Result<()> {
    let (_watcher, watch_rx) = setup_watcher(&app.source, app.watch)?;

    while !app.should_quit {
        terminal.draw(|frame| draw(frame, app))?;

        if let Some(rx) = watch_rx.as_ref() {
            for event in rx.try_iter() {
                if event.is_ok() {
                    app.reload();
                }
            }
        }

        if event::poll(Duration::from_millis(100))? {
            match event::read()? {
                Event::Key(key) => {
                    let height = terminal.size()?.height.saturating_sub(3) as usize;
                    app.handle_key(key, height.max(1));
                }
                Event::Resize(width, _) => {
                    app.rerender(width.saturating_sub(4) as usize);
                }
                _ => {}
            }
        }
    }

    Ok(())
}

fn draw(frame: &mut Frame, app: &mut PreviewApp) {
    let area = frame.area();
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([Constraint::Min(1), Constraint::Length(1)])
        .split(area);

    app.rerender(chunks[0].width.saturating_sub(4) as usize);

    let visible_height = chunks[0].height.saturating_sub(2) as usize;
    let visible = app
        .rendered
        .lines
        .iter()
        .skip(app.scroll)
        .take(visible_height)
        .map(|line| line_to_tui(line, &app.theme, Some(&app.search_query)))
        .collect::<Vec<_>>();

    let title = if app.watch {
        format!(" mdview - {} - live ", source_name(&app.source))
    } else {
        format!(" mdview - {} ", source_name(&app.source))
    };

    let document = Paragraph::new(visible)
        .block(
            TuiBlock::default()
                .borders(Borders::ALL)
                .border_style(app.theme.border_style())
                .title(title),
        )
        .wrap(Wrap { trim: false });
    frame.render_widget(document, chunks[0]);

    let status_style = if app.error.is_some() {
        Style::default().fg(Color::Red).add_modifier(Modifier::BOLD)
    } else {
        app.theme.muted_style()
    };
    frame.render_widget(
        Paragraph::new(Line::from(Span::styled(app.status_text(), status_style))),
        chunks[1],
    );

    if app.outline_open {
        draw_outline(frame, app, centered_rect(70, 70, area));
    }
}

fn draw_outline(frame: &mut Frame, app: &PreviewApp, area: Rect) {
    let items = if app.document.headings.is_empty() {
        vec![TuiListItem::new(Line::from("No headings"))]
    } else {
        app.document
            .headings
            .iter()
            .enumerate()
            .map(|(index, heading)| {
                let prefix = "  ".repeat(heading.level.saturating_sub(1) as usize);
                let marker = if index == app.outline_index {
                    "> "
                } else {
                    "  "
                };
                let style = if index == app.outline_index {
                    app.theme.heading_style()
                } else {
                    app.theme.body_style()
                };
                TuiListItem::new(Line::from(Span::styled(
                    format!("{marker}{prefix}{}", heading.text),
                    style,
                )))
            })
            .collect()
    };

    let list = List::new(items).block(
        TuiBlock::default()
            .title(" outline ")
            .borders(Borders::ALL)
            .border_style(app.theme.border_style()),
    );
    frame.render_widget(Clear, area);
    frame.render_widget(list, area);
}

fn centered_rect(percent_x: u16, percent_y: u16, area: Rect) -> Rect {
    let vertical = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage((100 - percent_y) / 2),
            Constraint::Percentage(percent_y),
            Constraint::Percentage((100 - percent_y) / 2),
        ])
        .split(area);

    Layout::default()
        .direction(Direction::Horizontal)
        .constraints([
            Constraint::Percentage((100 - percent_x) / 2),
            Constraint::Percentage(percent_x),
            Constraint::Percentage((100 - percent_x) / 2),
        ])
        .split(vertical[1])[1]
}

fn setup_watcher(
    source: &Source,
    enabled: bool,
) -> Result<(
    Option<RecommendedWatcher>,
    Option<Receiver<notify::Result<notify::Event>>>,
)> {
    if !enabled {
        return Ok((None, None));
    }

    let Source::File(path) = source else {
        return Ok((None, None));
    };

    let (tx, rx) = mpsc::channel();
    let mut watcher = RecommendedWatcher::new(
        move |event| {
            let _ = tx.send(event);
        },
        Config::default(),
    )?;
    watcher
        .watch(Path::new(path), RecursiveMode::NonRecursive)
        .with_context(|| format!("failed to watch {}", path.display()))?;
    Ok((Some(watcher), Some(rx)))
}
