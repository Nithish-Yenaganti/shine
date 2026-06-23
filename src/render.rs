use base64::{engine::general_purpose, Engine as _};
use ratatui::{
    style::{Color, Modifier, Style},
    text::{Line, Span},
};
use syntect::{easy::HighlightLines, highlighting::ThemeSet, parsing::SyntaxSet};
use unicode_width::UnicodeWidthStr;

use crate::{
    model::{Block, CalloutKind, Document, ListItem, RichText},
    theme::{parse_color, Theme},
};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum LineKind {
    Heading(u8),
    Body,
    Muted,
    Code,
    Quote,
    Callout(CalloutKind),
    Table,
    Divider,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum SegmentRole {
    Normal,
    Bold,
    Italic,
    InlineCode,
    Link,
    Muted,
    TableHeader,
    CodeLineNumber,
    Syntax,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct RenderedLine {
    pub text: String,
    pub kind: LineKind,
    pub block_index: Option<usize>,
    pub spans: Option<Vec<StyledSegment>>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct StyledSegment {
    pub text: String,
    pub fg: Option<(u8, u8, u8)>,
    pub role: SegmentRole,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct RenderedDocument {
    pub lines: Vec<RenderedLine>,
    pub block_line_offsets: Vec<usize>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct RenderOptions {
    pub block_gap: usize,
    pub code_line_numbers: bool,
}

impl Default for RenderOptions {
    fn default() -> Self {
        Self {
            block_gap: 1,
            code_line_numbers: true,
        }
    }
}

pub fn render_document(document: &Document, width: usize) -> RenderedDocument {
    render_document_with_options(document, width, RenderOptions::default())
}

pub fn render_document_with_theme(
    document: &Document,
    width: usize,
    theme: &Theme,
) -> RenderedDocument {
    render_document_with_options(
        document,
        width,
        RenderOptions {
            block_gap: theme.block_padding.unwrap_or(1) as usize,
            code_line_numbers: theme.code_line_numbers.unwrap_or(true),
        },
    )
}

pub fn render_document_with_options(
    document: &Document,
    width: usize,
    options: RenderOptions,
) -> RenderedDocument {
    let width = width.max(28);
    let mut lines = Vec::new();
    let mut block_line_offsets = Vec::with_capacity(document.blocks.len());

    for (block_index, block) in document.blocks.iter().enumerate() {
        block_line_offsets.push(lines.len());
        match block {
            Block::Heading { level, content } => {
                render_heading(&mut lines, block_index, *level, content, width)
            }
            Block::Paragraph { content } => {
                push_rich_wrapped(
                    &mut lines,
                    Some(block_index),
                    content,
                    width,
                    "",
                    LineKind::Body,
                );
                push_blank(&mut lines);
            }
            Block::CodeBlock { language, code } => {
                render_code(&mut lines, block_index, language.as_deref(), code, width);
            }
            Block::Quote { content } => {
                render_quote(&mut lines, block_index, content, width);
            }
            Block::Callout { kind, title, body } => {
                render_callout(&mut lines, block_index, *kind, title, body, width);
            }
            Block::List { ordered, items } => {
                render_list(&mut lines, block_index, *ordered, items, width);
            }
            Block::Table { headers, rows } => {
                render_table(&mut lines, block_index, headers, rows, width);
            }
            Block::Divider => {
                push_line(
                    &mut lines,
                    "─".repeat(width.min(72)),
                    LineKind::Divider,
                    Some(block_index),
                    None,
                );
                push_blank(&mut lines);
            }
            Block::Image { alt, url } => {
                render_image(&mut lines, block_index, alt, url, width);
            }
        }
    }

    let mut rendered = RenderedDocument {
        lines,
        block_line_offsets,
    };
    apply_render_options(&mut rendered, options);
    rendered
}

pub fn plain_text(rendered: &RenderedDocument) -> String {
    let mut output = rendered
        .lines
        .iter()
        .map(|line| line.text.as_str())
        .collect::<Vec<_>>()
        .join("\n");
    output.push('\n');
    output
}

pub fn line_to_tui<'a>(line: &'a RenderedLine, theme: &Theme, query: Option<&str>) -> Line<'a> {
    let base = style_for_kind(line.kind, theme);
    let Some(query) = query.filter(|query| !query.trim().is_empty()) else {
        if let Some(spans) = &line.spans {
            return Line::from(
                spans
                    .iter()
                    .map(|segment| {
                        Span::styled(segment.text.clone(), segment_style(segment, base, theme))
                    })
                    .collect::<Vec<_>>(),
            );
        }
        return Line::from(Span::styled(line.text.clone(), base));
    };

    let lower_text = line.text.to_ascii_lowercase();
    let lower_query = query.to_ascii_lowercase();
    let mut spans = Vec::new();
    let mut cursor = 0;

    for (start, _) in lower_text.match_indices(&lower_query) {
        if start > cursor {
            spans.push(Span::styled(line.text[cursor..start].to_string(), base));
        }
        let end = start + query.len();
        spans.push(Span::styled(
            line.text[start..end].to_string(),
            theme.match_style(),
        ));
        cursor = end;
    }

    if cursor < line.text.len() {
        spans.push(Span::styled(line.text[cursor..].to_string(), base));
    }

    Line::from(spans)
}

pub fn style_for_kind(kind: LineKind, theme: &Theme) -> Style {
    match kind {
        LineKind::Heading(_) => theme.heading_style(),
        LineKind::Body => theme.body_style(),
        LineKind::Muted => theme.muted_style(),
        LineKind::Code => theme.code_style(),
        LineKind::Quote => theme.quote_style(),
        LineKind::Callout(CalloutKind::Warning) | LineKind::Callout(CalloutKind::Caution) => {
            Style::default().fg(parse_color(&theme.callout_warning))
        }
        LineKind::Callout(CalloutKind::Tip) => Style::default().fg(parse_color(&theme.callout_tip)),
        LineKind::Callout(_) => Style::default().fg(parse_color(&theme.callout_note)),
        LineKind::Table => theme.body_style(),
        LineKind::Divider => theme.border_style(),
    }
}

fn segment_style(segment: &StyledSegment, base: Style, theme: &Theme) -> Style {
    let mut style = base;
    if let Some((r, g, b)) = segment.fg {
        style = style.fg(Color::Rgb(r, g, b));
    }

    match segment.role {
        SegmentRole::Normal => style,
        SegmentRole::Bold => style.add_modifier(Modifier::BOLD),
        SegmentRole::Italic => style.add_modifier(Modifier::ITALIC),
        SegmentRole::InlineCode => theme.inline_code_style(),
        SegmentRole::Link => theme.link_style(),
        SegmentRole::Muted => theme.muted_style(),
        SegmentRole::TableHeader => Style::default()
            .fg(parse_color(&theme.table_header))
            .add_modifier(Modifier::BOLD),
        SegmentRole::CodeLineNumber => Style::default().fg(parse_color(&theme.code_line_number)),
        SegmentRole::Syntax => style,
    }
}

fn apply_render_options(rendered: &mut RenderedDocument, options: RenderOptions) {
    let old_lines = std::mem::take(&mut rendered.lines);
    let mut new_lines = Vec::with_capacity(old_lines.len());
    let mut index_map = vec![0usize; old_lines.len() + 1];

    for (old_index, mut line) in old_lines.into_iter().enumerate() {
        index_map[old_index] = new_lines.len();

        if line.text.is_empty() {
            for _ in 0..options.block_gap {
                new_lines.push(line.clone());
            }
            continue;
        }

        if !options.code_line_numbers && matches!(line.kind, LineKind::Code) {
            strip_code_line_number(&mut line);
        }

        new_lines.push(line);
    }

    index_map
        .last_mut()
        .map(|last| *last = new_lines.len())
        .unwrap_or_default();
    rendered.block_line_offsets = rendered
        .block_line_offsets
        .iter()
        .map(|offset| index_map.get(*offset).copied().unwrap_or(new_lines.len()))
        .collect();
    rendered.lines = new_lines;
}

fn strip_code_line_number(line: &mut RenderedLine) {
    const GUTTER_LEN: usize = 7;
    if !line.text.starts_with("│ ") || line.text.len() <= GUTTER_LEN {
        return;
    }

    let code = line.text[GUTTER_LEN..].to_string();
    line.text = format!("│ {code}");
    if let Some(spans) = line.spans.as_mut() {
        if !spans.is_empty() {
            spans[0] = styled("│ ", SegmentRole::Muted);
        }
    }
}

fn render_heading(
    lines: &mut Vec<RenderedLine>,
    block_index: usize,
    level: u8,
    content: &RichText,
    width: usize,
) {
    if !lines.is_empty() {
        push_blank(lines);
    }

    let text = content.normalized_plain_text();
    let label = match level {
        1 => text,
        2 => text,
        _ => format!("{} {text}", ".".repeat(level.saturating_sub(2) as usize)),
    };

    push_line(
        lines,
        label.clone(),
        LineKind::Heading(level),
        Some(block_index),
        Some(vec![StyledSegment {
            text: label.clone(),
            fg: None,
            role: SegmentRole::Bold,
        }]),
    );

    if level <= 2 {
        let underline_width = UnicodeWidthStr::width(label.as_str()).clamp(3, width.min(72));
        push_line(
            lines,
            if level == 1 { "═" } else { "─" }.repeat(underline_width),
            LineKind::Divider,
            Some(block_index),
            None,
        );
    }
    push_blank(lines);
}

fn render_quote(
    lines: &mut Vec<RenderedLine>,
    block_index: usize,
    content: &RichText,
    width: usize,
) {
    push_rich_wrapped(
        lines,
        Some(block_index),
        content,
        width,
        "│ ",
        LineKind::Quote,
    );
    push_blank(lines);
}

fn render_code(
    lines: &mut Vec<RenderedLine>,
    block_index: usize,
    language: Option<&str>,
    code: &str,
    width: usize,
) {
    let syntax_set = SyntaxSet::load_defaults_newlines();
    let theme_set = ThemeSet::load_defaults();
    let syntax = language
        .and_then(|language| syntax_set.find_syntax_by_token(language))
        .unwrap_or_else(|| syntax_set.find_syntax_plain_text());
    let mut highlighter = theme_set
        .themes
        .get("base16-ocean.dark")
        .or_else(|| theme_set.themes.values().next())
        .map(|theme| HighlightLines::new(syntax, theme));

    let label = language
        .filter(|language| !language.trim().is_empty())
        .unwrap_or("plain");
    let header = format!("┌─ {label}");
    push_line(lines, header, LineKind::Divider, Some(block_index), None);

    for (index, raw_line) in code.lines().enumerate() {
        push_code_line(
            lines,
            block_index,
            index + 1,
            raw_line,
            width,
            &syntax_set,
            highlighter.as_mut(),
        );
    }
    push_line(
        lines,
        "└".to_string() + &"─".repeat(width.saturating_sub(1).min(72)),
        LineKind::Divider,
        Some(block_index),
        None,
    );
    push_blank(lines);
}

fn render_callout(
    lines: &mut Vec<RenderedLine>,
    block_index: usize,
    kind: CalloutKind,
    title: &str,
    body: &RichText,
    width: usize,
) {
    let title = if title.trim().is_empty() {
        kind.label()
    } else {
        title.trim()
    };
    let label = format!("{}  {title}", kind.label());
    push_line(
        lines,
        format!("┃ {label}"),
        LineKind::Callout(kind),
        Some(block_index),
        Some(vec![
            styled("┃ ", SegmentRole::Muted),
            styled(kind.label(), SegmentRole::Bold),
            styled(format!("  {title}"), SegmentRole::Normal),
        ]),
    );

    if !body.is_empty() {
        push_rich_wrapped(
            lines,
            Some(block_index),
            body,
            width.saturating_sub(4),
            "┃ ",
            LineKind::Callout(kind),
        );
    }
    push_blank(lines);
}

fn render_list(
    lines: &mut Vec<RenderedLine>,
    block_index: usize,
    ordered: bool,
    items: &[ListItem],
    width: usize,
) {
    for (index, item) in items.iter().enumerate() {
        let marker = if ordered {
            format!("{}.", index + 1)
        } else if let Some(checked) = item.checked {
            if checked {
                "[x]".to_string()
            } else {
                "[ ]".to_string()
            }
        } else {
            "-".to_string()
        };
        push_rich_wrapped(
            lines,
            Some(block_index),
            &item.content,
            width,
            &format!("{marker} "),
            LineKind::Body,
        );
    }
    push_blank(lines);
}

fn render_table(
    lines: &mut Vec<RenderedLine>,
    block_index: usize,
    headers: &[String],
    rows: &[Vec<String>],
    width: usize,
) {
    if headers.is_empty() {
        return;
    }

    let column_count = headers
        .len()
        .max(rows.iter().map(Vec::len).max().unwrap_or_default());
    let mut widths = vec![3usize; column_count];

    for (index, header) in headers.iter().enumerate() {
        widths[index] = widths[index].max(UnicodeWidthStr::width(header.as_str()).min(28));
    }
    for row in rows {
        for (index, cell) in row.iter().enumerate() {
            widths[index] = widths[index].max(UnicodeWidthStr::width(cell.as_str()).min(28));
        }
    }

    let total_padding = 3 * column_count + 1;
    let max_content = width.saturating_sub(total_padding).max(column_count * 4);
    while widths.iter().sum::<usize>() > max_content {
        if let Some((index, _)) = widths.iter().enumerate().max_by_key(|(_, value)| *value) {
            if widths[index] <= 4 {
                break;
            }
            widths[index] -= 1;
        }
    }

    let top_border = table_border(&widths, BorderPosition::Top);
    let mid_border = table_border(&widths, BorderPosition::Middle);
    let bottom_border = table_border(&widths, BorderPosition::Bottom);
    push_line(
        lines,
        top_border,
        LineKind::Divider,
        Some(block_index),
        None,
    );
    push_table_row(lines, block_index, headers, &widths, true);
    push_line(
        lines,
        mid_border,
        LineKind::Divider,
        Some(block_index),
        None,
    );
    for row in rows {
        push_table_row(lines, block_index, row, &widths, false);
    }
    push_line(
        lines,
        bottom_border,
        LineKind::Divider,
        Some(block_index),
        None,
    );
    push_blank(lines);
}

fn render_image(
    lines: &mut Vec<RenderedLine>,
    block_index: usize,
    alt: &str,
    url: &str,
    _width: usize,
) {
    let protocol = terminal_image_protocol();
    let label = if alt.trim().is_empty() { "image" } else { alt };
    let status = protocol
        .map(|protocol| format!("{protocol} image preview capable"))
        .unwrap_or_else(|| "image preview unavailable in this terminal".to_string());

    push_line(
        lines,
        format!("image  {label}"),
        LineKind::Muted,
        Some(block_index),
        Some(vec![
            styled("image  ", SegmentRole::Muted),
            styled(label, SegmentRole::Bold),
        ]),
    );
    push_line(
        lines,
        format!("       {url}"),
        LineKind::Muted,
        Some(block_index),
        Some(vec![
            styled("       ", SegmentRole::Muted),
            styled(url, SegmentRole::Link),
        ]),
    );
    push_line(
        lines,
        format!("       {status}"),
        LineKind::Muted,
        Some(block_index),
        None,
    );
    if let Some(protocol) = protocol {
        if let Some(escape) = terminal_image_escape(protocol, url) {
            push_line(
                lines,
                escape,
                LineKind::Muted,
                Some(block_index),
                Some(vec![styled(
                    format!("       rendered via {protocol} image protocol"),
                    SegmentRole::Muted,
                )]),
            );
        }
    }
    push_blank(lines);
}

fn push_code_line(
    lines: &mut Vec<RenderedLine>,
    block_index: usize,
    number: usize,
    raw_line: &str,
    width: usize,
    syntax_set: &SyntaxSet,
    highlighter: Option<&mut HighlightLines<'_>>,
) {
    let gutter = format!("│ {:>3}  ", number);
    let available = width
        .saturating_sub(UnicodeWidthStr::width(gutter.as_str()))
        .max(8);
    let code = clip_width(raw_line, available);
    let text = format!("{gutter}{code}");
    let mut spans = vec![StyledSegment {
        text: gutter,
        fg: None,
        role: SegmentRole::CodeLineNumber,
    }];

    if let Some(highlighter) = highlighter {
        match highlighter.highlight_line(&code, syntax_set) {
            Ok(ranges) => {
                spans.extend(ranges.into_iter().map(|(style, token)| StyledSegment {
                    text: token.to_string(),
                    fg: Some((style.foreground.r, style.foreground.g, style.foreground.b)),
                    role: SegmentRole::Syntax,
                }));
            }
            Err(_) => spans.push(styled(code, SegmentRole::Normal)),
        }
    } else {
        spans.push(styled(code, SegmentRole::Normal));
    }

    push_line(lines, text, LineKind::Code, Some(block_index), Some(spans));
}

fn push_rich_wrapped(
    lines: &mut Vec<RenderedLine>,
    block_index: Option<usize>,
    content: &RichText,
    width: usize,
    prefix: &str,
    kind: LineKind,
) {
    let available = width.saturating_sub(UnicodeWidthStr::width(prefix)).max(8);
    let words = rich_words(content);

    if words.is_empty() {
        push_line(lines, prefix.trim_end(), kind, block_index, None);
        return;
    }

    let mut current = Vec::new();
    let mut current_width = 0usize;

    for word in words {
        let word_width = segments_width(&word);
        let separator = usize::from(current_width > 0 && !sticks_to_previous(&word));
        if current_width > 0 && current_width + separator + word_width > available {
            push_rich_line(
                lines,
                block_index,
                prefix,
                kind,
                std::mem::take(&mut current),
            );
            current_width = 0;
        }

        if current_width > 0 && !sticks_to_previous(&word) {
            current.push(styled(" ", SegmentRole::Normal));
            current_width += 1;
        }
        current_width += word_width;
        current.extend(word);
    }

    if !current.is_empty() {
        push_rich_line(lines, block_index, prefix, kind, current);
    }
}

fn push_rich_line(
    lines: &mut Vec<RenderedLine>,
    block_index: Option<usize>,
    prefix: &str,
    kind: LineKind,
    mut spans: Vec<StyledSegment>,
) {
    if !prefix.is_empty() {
        spans.insert(0, styled(prefix, SegmentRole::Muted));
    }
    let text = spans
        .iter()
        .map(|span| span.text.as_str())
        .collect::<String>();
    push_line(lines, text, kind, block_index, Some(spans));
}

fn rich_words(content: &RichText) -> Vec<Vec<StyledSegment>> {
    let mut words = Vec::new();
    for span in &content.spans {
        if span.code {
            words.push(vec![styled(span.text.trim(), SegmentRole::InlineCode)]);
            continue;
        }

        for part in span.text.split_whitespace() {
            let mut role = SegmentRole::Normal;
            if span.link.is_some() {
                role = SegmentRole::Link;
            } else if span.bold {
                role = SegmentRole::Bold;
            } else if span.italic {
                role = SegmentRole::Italic;
            }

            let text = if span.link.is_some() {
                format!("{part}↗")
            } else {
                part.to_string()
            };
            words.push(vec![styled(text, role)]);
        }
    }
    words
}

fn push_table_row(
    lines: &mut Vec<RenderedLine>,
    block_index: usize,
    cells: &[String],
    widths: &[usize],
    header: bool,
) {
    let wrapped_cells = widths
        .iter()
        .enumerate()
        .map(|(index, width)| {
            let cell = cells.get(index).map(String::as_str).unwrap_or("");
            textwrap::wrap(cell, *width)
                .into_iter()
                .map(|line| line.to_string())
                .collect::<Vec<_>>()
        })
        .collect::<Vec<_>>();
    let height = wrapped_cells.iter().map(Vec::len).max().unwrap_or(1).max(1);

    for line_index in 0..height {
        let mut text = String::from("│");
        let mut spans = vec![styled("│", SegmentRole::Muted)];
        for (cell_index, width) in widths.iter().enumerate() {
            let cell = wrapped_cells
                .get(cell_index)
                .and_then(|lines| lines.get(line_index))
                .map(String::as_str)
                .unwrap_or("");
            let padding = width.saturating_sub(UnicodeWidthStr::width(cell));
            let role = if header {
                SegmentRole::TableHeader
            } else {
                SegmentRole::Normal
            };
            text.push(' ');
            text.push_str(cell);
            text.push_str(&" ".repeat(padding + 1));
            text.push('│');
            spans.push(styled(" ", SegmentRole::Muted));
            spans.push(styled(cell, role));
            spans.push(styled(" ".repeat(padding + 1), SegmentRole::Muted));
            spans.push(styled("│", SegmentRole::Muted));
        }
        push_line(lines, text, LineKind::Table, Some(block_index), Some(spans));
    }
}

fn push_line(
    lines: &mut Vec<RenderedLine>,
    text: impl Into<String>,
    kind: LineKind,
    block_index: Option<usize>,
    spans: Option<Vec<StyledSegment>>,
) {
    lines.push(RenderedLine {
        text: text.into(),
        kind,
        block_index,
        spans,
    });
}

fn push_blank(lines: &mut Vec<RenderedLine>) {
    if !matches!(lines.last(), Some(line) if line.text.is_empty()) {
        push_line(lines, "", LineKind::Muted, None, None);
    }
}

fn styled(text: impl Into<String>, role: SegmentRole) -> StyledSegment {
    StyledSegment {
        text: text.into(),
        fg: None,
        role,
    }
}

fn segments_width(segments: &[StyledSegment]) -> usize {
    segments
        .iter()
        .map(|segment| UnicodeWidthStr::width(segment.text.as_str()))
        .sum()
}

fn sticks_to_previous(segments: &[StyledSegment]) -> bool {
    let Some(first) = segments.first() else {
        return false;
    };
    first
        .text
        .chars()
        .next()
        .is_some_and(|ch| matches!(ch, ',' | '.' | ';' | ':' | '!' | '?' | ')' | ']' | '}'))
}

#[derive(Debug, Clone, Copy)]
enum BorderPosition {
    Top,
    Middle,
    Bottom,
}

fn table_border(widths: &[usize], position: BorderPosition) -> String {
    let (left, join, right) = match position {
        BorderPosition::Top => ('┌', '┬', '┐'),
        BorderPosition::Middle => ('├', '┼', '┤'),
        BorderPosition::Bottom => ('└', '┴', '┘'),
    };
    let mut out = String::from(left);
    for width in widths {
        out.push_str(&"─".repeat(width + 2));
        out.push(join);
    }
    out.pop();
    out.push(right);
    out
}

fn clip_width(value: &str, width: usize) -> String {
    if UnicodeWidthStr::width(value) <= width {
        return value.to_string();
    }

    let mut out = String::new();
    let mut used = 0;
    for ch in value.chars() {
        let ch_width = UnicodeWidthStr::width(ch.to_string().as_str());
        if used + ch_width + 1 > width {
            break;
        }
        out.push(ch);
        used += ch_width;
    }
    out.push('.');
    out
}

fn terminal_image_protocol() -> Option<&'static str> {
    let term = std::env::var("TERM")
        .unwrap_or_default()
        .to_ascii_lowercase();
    let term_program = std::env::var("TERM_PROGRAM")
        .unwrap_or_default()
        .to_ascii_lowercase();

    if term.contains("kitty") {
        Some("kitty")
    } else if term_program.contains("iterm") {
        Some("iterm2")
    } else if term.contains("sixel") || std::env::var("TERM_IMAGE_PROTOCOL").is_ok() {
        Some("sixel")
    } else {
        None
    }
}

fn terminal_image_escape(protocol: &str, url: &str) -> Option<String> {
    let path = std::path::Path::new(url);
    if !path.exists() || !path.is_file() {
        return None;
    }

    match protocol {
        "kitty" => {
            let path = path.canonicalize().ok()?;
            let encoded_path = general_purpose::STANDARD.encode(path.to_string_lossy().as_bytes());
            Some(format!("\x1b_Ga=T,t=f,f=100;{encoded_path}\x1b\\"))
        }
        "iterm2" => {
            let bytes = std::fs::read(path).ok()?;
            let encoded = general_purpose::STANDARD.encode(bytes);
            Some(format!(
                "\x1b]1337;File=inline=1;preserveAspectRatio=1:{encoded}\x07"
            ))
        }
        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use std::io::Write;

    use tempfile::NamedTempFile;

    use super::terminal_image_escape;

    #[test]
    fn builds_kitty_image_escape_for_local_files() {
        let mut file = NamedTempFile::new().expect("temp image");
        file.write_all(b"fake image bytes")
            .expect("write temp image");
        let escape = terminal_image_escape("kitty", file.path().to_str().expect("path"))
            .expect("kitty escape");

        assert!(escape.starts_with("\x1b_G"));
        assert!(escape.ends_with("\x1b\\"));
    }
}
