use mdview::{
    parser::parse_document,
    render::{plain_text, render_document, render_document_with_theme, SegmentRole},
    theme::Theme,
};
use unicode_width::UnicodeWidthStr;

#[test]
fn renders_basic_markdown_without_raw_heading_syntax() {
    let document = parse_document(include_str!("../fixtures/basic.md"), None);
    let rendered = render_document(&document, 72);
    let output = plain_text(&rendered);

    assert!(output.contains("mdview"));
    assert!(output.contains("Features"));
    assert!(output.contains("┃ NOTE  NOTE"));
    assert!(output.contains("┌─ rust"));
    assert!(output.contains("│   1  fn main()"));
    assert!(output.contains("image  screenshot"));
    assert!(!output.contains("# mdview"));
}

#[test]
fn renders_width_constrained_tables() {
    let document = parse_document(
        "| Very long heading | Status |\n| --- | --- |\n| very long value that should clip | ok |\n",
        None,
    );
    let rendered = render_document(&document, 36);

    assert!(
        rendered
            .lines
            .iter()
            .all(|line| UnicodeWidthStr::width(line.text.as_str()) <= 80),
        "renderer should not produce runaway table lines"
    );
}

#[test]
fn renders_code_lines_with_syntax_spans() {
    let document = parse_document("```rust\nfn main() {}\n```\n", None);
    let rendered = render_document(&document, 72);

    assert!(
        rendered
            .lines
            .iter()
            .any(|line| line.spans.as_ref().is_some_and(|spans| spans.len() > 1)),
        "expected code line to include syntax-highlighted spans"
    );
}

#[test]
fn preserves_mixed_inline_styles_in_paragraph_lines() {
    let document = parse_document(
        "This is **bold**, *italic*, `code`, and [a link](https://example.com).\n",
        None,
    );
    let rendered = render_document(&document, 120);
    let roles = rendered
        .lines
        .iter()
        .flat_map(|line| line.spans.as_deref().unwrap_or_default())
        .map(|span| span.role)
        .collect::<Vec<_>>();

    assert!(roles.contains(&SegmentRole::Bold));
    assert!(roles.contains(&SegmentRole::Italic));
    assert!(roles.contains(&SegmentRole::InlineCode));
    assert!(roles.contains(&SegmentRole::Link));
}

#[test]
fn wraps_table_cells_instead_of_only_clipping() {
    let document = parse_document(
        "| Feature | Description |\n| --- | --- |\n| Layout | wraps long cell content into readable rows |\n",
        None,
    );
    let rendered = render_document(&document, 44);
    let output = plain_text(&rendered);

    assert!(output.contains("readable"));
    assert!(output.contains("rows"));
}

#[test]
fn theme_controls_spacing_and_code_line_numbers() {
    let document = parse_document("# Title\n\n```rust\nfn main() {}\n```\n", None);
    let mut theme = Theme::default();
    theme.block_padding = Some(0);
    theme.code_line_numbers = Some(false);

    let rendered = render_document_with_theme(&document, 72, &theme);
    let output = plain_text(&rendered);

    assert!(!output.contains("│   1  fn main"));
    assert!(output.contains("│   fn main"));
    assert!(!output.contains("\n\n\n"));
}
