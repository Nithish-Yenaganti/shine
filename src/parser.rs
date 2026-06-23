use pulldown_cmark::{CodeBlockKind, Event, HeadingLevel, Options, Parser, Tag};

use crate::model::{
    normalize_plain_text, Block, CalloutKind, Document, InlineSpan, ListItem, RichText,
};

#[derive(Debug, Clone, Default)]
struct InlineStyle {
    bold: bool,
    italic: bool,
    code: bool,
    link: Option<String>,
}

#[derive(Debug, Clone, Default)]
struct InlineBuilder {
    spans: Vec<InlineSpan>,
}

impl InlineBuilder {
    fn push(&mut self, text: &str, style: &InlineStyle) {
        if text.is_empty() {
            return;
        }

        if let Some(last) = self.spans.last_mut() {
            if last.bold == style.bold
                && last.italic == style.italic
                && last.code == style.code
                && last.link == style.link
            {
                last.text.push_str(text);
                return;
            }
        }

        self.spans.push(InlineSpan {
            text: text.to_string(),
            bold: style.bold,
            italic: style.italic,
            code: style.code,
            link: style.link.clone(),
        });
    }

    fn build(self) -> RichText {
        RichText {
            spans: self
                .spans
                .into_iter()
                .filter(|span| !span.text.is_empty())
                .collect(),
        }
    }
}

#[derive(Debug, Default)]
struct ListBuilder {
    ordered: bool,
    items: Vec<ListItem>,
    current_content: InlineBuilder,
    current_checked: Option<bool>,
}

#[derive(Debug, Default)]
struct TableBuilder {
    headers: Vec<String>,
    rows: Vec<Vec<String>>,
    current_row: Vec<String>,
    current_cell: String,
    in_head: bool,
}

pub fn parse_document(source: &str, source_name: Option<String>) -> Document {
    let mut options = Options::empty();
    options.insert(Options::ENABLE_TABLES);
    options.insert(Options::ENABLE_TASKLISTS);
    options.insert(Options::ENABLE_STRIKETHROUGH);
    options.insert(Options::ENABLE_FOOTNOTES);

    let parser = Parser::new_ext(source, options);
    let mut blocks = Vec::new();

    let mut style = InlineStyle::default();
    let mut paragraph: Option<InlineBuilder> = None;
    let mut heading: Option<(u8, InlineBuilder)> = None;
    let mut code: Option<(Option<String>, String)> = None;
    let mut quote: Option<InlineBuilder> = None;
    let mut list: Option<ListBuilder> = None;
    let mut table: Option<TableBuilder> = None;
    let mut image: Option<(String, String)> = None;

    for event in parser {
        match event {
            Event::Start(tag) => match tag {
                Tag::Paragraph => {
                    if list.is_none() && quote.is_none() && table.is_none() && image.is_none() {
                        paragraph = Some(InlineBuilder::default());
                    }
                }
                Tag::Heading(level, _, _) => {
                    heading = Some((heading_level(level), InlineBuilder::default()));
                }
                Tag::Emphasis => style.italic = true,
                Tag::Strong => style.bold = true,
                Tag::Link(_, url, _) => style.link = Some(url.to_string()),
                Tag::CodeBlock(kind) => {
                    let language = match kind {
                        CodeBlockKind::Fenced(language) if !language.is_empty() => {
                            Some(language.to_string())
                        }
                        _ => None,
                    };
                    code = Some((language, String::new()));
                }
                Tag::BlockQuote => {
                    quote = Some(InlineBuilder::default());
                }
                Tag::List(start) => {
                    list = Some(ListBuilder {
                        ordered: start.is_some(),
                        ..ListBuilder::default()
                    });
                }
                Tag::Item => {
                    if let Some(list) = list.as_mut() {
                        list.current_content = InlineBuilder::default();
                        list.current_checked = None;
                    }
                }
                Tag::Table(_) => {
                    table = Some(TableBuilder::default());
                }
                Tag::TableHead => {
                    if let Some(table) = table.as_mut() {
                        table.in_head = true;
                        table.current_row.clear();
                    }
                }
                Tag::TableRow => {
                    if let Some(table) = table.as_mut() {
                        table.current_row.clear();
                    }
                }
                Tag::TableCell => {
                    if let Some(table) = table.as_mut() {
                        table.current_cell.clear();
                    }
                }
                Tag::Image(_, url, _) => {
                    image = Some((String::new(), url.to_string()));
                }
                _ => {}
            },
            Event::End(tag) => match tag {
                Tag::Paragraph => {
                    if let Some(content) = paragraph.take() {
                        push_rich_block(
                            &mut blocks,
                            Block::Paragraph {
                                content: content.build(),
                            },
                        );
                    }
                }
                Tag::Heading(_, _, _) => {
                    if let Some((level, content)) = heading.take() {
                        push_rich_block(
                            &mut blocks,
                            Block::Heading {
                                level,
                                content: content.build(),
                            },
                        );
                    }
                }
                Tag::Emphasis => style.italic = false,
                Tag::Strong => style.bold = false,
                Tag::Link(_, _, _) => style.link = None,
                Tag::CodeBlock(_) => {
                    if let Some((language, code)) = code.take() {
                        blocks.push(Block::CodeBlock { language, code });
                    }
                }
                Tag::BlockQuote => {
                    if let Some(content) = quote.take() {
                        blocks.push(parse_quote_or_callout(content.build()));
                    }
                }
                Tag::Item => {
                    if let Some(list) = list.as_mut() {
                        let content = normalize_rich_text(list.current_content.clone().build());
                        if !content.is_empty() || list.current_checked.is_some() {
                            list.items.push(ListItem {
                                content,
                                checked: list.current_checked,
                            });
                        }
                    }
                }
                Tag::List(_) => {
                    if let Some(list) = list.take() {
                        blocks.push(Block::List {
                            ordered: list.ordered,
                            items: list.items,
                        });
                    }
                }
                Tag::TableCell => {
                    if let Some(table) = table.as_mut() {
                        table
                            .current_row
                            .push(normalize_plain_text(&table.current_cell));
                    }
                }
                Tag::TableRow => {
                    if let Some(table) = table.as_mut() {
                        if table.in_head && table.headers.is_empty() {
                            table.headers = table.current_row.clone();
                        } else if !table.current_row.is_empty() {
                            table.rows.push(table.current_row.clone());
                        }
                    }
                }
                Tag::TableHead => {
                    if let Some(table) = table.as_mut() {
                        if table.headers.is_empty() && !table.current_row.is_empty() {
                            table.headers = table.current_row.clone();
                        }
                        table.in_head = false;
                    }
                }
                Tag::Table(_) => {
                    if let Some(table) = table.take() {
                        blocks.push(Block::Table {
                            headers: table.headers,
                            rows: table.rows,
                        });
                    }
                }
                Tag::Image(_, _, _) => {
                    if let Some((alt, url)) = image.take() {
                        blocks.push(Block::Image {
                            alt: normalize_plain_text(&alt),
                            url,
                        });
                    }
                }
                _ => {}
            },
            Event::Text(text) => append_text(
                &text,
                &style,
                &mut paragraph,
                &mut heading,
                &mut code,
                &mut quote,
                &mut list,
                &mut table,
                &mut image,
            ),
            Event::Code(text) => {
                let mut code_style = style.clone();
                code_style.code = true;
                append_text(
                    &text,
                    &code_style,
                    &mut paragraph,
                    &mut heading,
                    &mut code,
                    &mut quote,
                    &mut list,
                    &mut table,
                    &mut image,
                );
            }
            Event::SoftBreak | Event::HardBreak => append_text(
                "\n",
                &style,
                &mut paragraph,
                &mut heading,
                &mut code,
                &mut quote,
                &mut list,
                &mut table,
                &mut image,
            ),
            Event::Rule => blocks.push(Block::Divider),
            Event::TaskListMarker(checked) => {
                if let Some(list) = list.as_mut() {
                    list.current_checked = Some(checked);
                }
            }
            _ => {}
        }
    }

    Document::new(source_name, blocks)
}

fn append_text(
    text: &str,
    style: &InlineStyle,
    paragraph: &mut Option<InlineBuilder>,
    heading: &mut Option<(u8, InlineBuilder)>,
    code: &mut Option<(Option<String>, String)>,
    quote: &mut Option<InlineBuilder>,
    list: &mut Option<ListBuilder>,
    table: &mut Option<TableBuilder>,
    image: &mut Option<(String, String)>,
) {
    if let Some((alt, _)) = image.as_mut() {
        alt.push_str(text);
    } else if let Some(table) = table.as_mut() {
        table.current_cell.push_str(text);
    } else if let Some((_, code)) = code.as_mut() {
        code.push_str(text);
    } else if let Some((_, heading)) = heading.as_mut() {
        heading.push(text, style);
    } else if let Some(list) = list.as_mut() {
        list.current_content.push(text, style);
    } else if let Some(quote) = quote.as_mut() {
        quote.push(text, style);
    } else if let Some(paragraph) = paragraph.as_mut() {
        paragraph.push(text, style);
    }
}

fn heading_level(level: HeadingLevel) -> u8 {
    match level {
        HeadingLevel::H1 => 1,
        HeadingLevel::H2 => 2,
        HeadingLevel::H3 => 3,
        HeadingLevel::H4 => 4,
        HeadingLevel::H5 => 5,
        HeadingLevel::H6 => 6,
    }
}

fn push_rich_block(blocks: &mut Vec<Block>, block: Block) {
    match block {
        Block::Heading { level, content } => {
            let content = normalize_rich_text(content);
            if !content.is_empty() {
                blocks.push(Block::Heading { level, content });
            }
        }
        Block::Paragraph { content } => {
            let content = normalize_rich_text(content);
            if !content.is_empty() {
                blocks.push(Block::Paragraph { content });
            }
        }
        other => blocks.push(other),
    }
}

fn parse_quote_or_callout(content: RichText) -> Block {
    let plain = content.plain_text();
    let normalized = plain.trim().to_string();
    let mut lines = normalized.lines();
    let first = lines.next().unwrap_or_default().trim();

    if first.starts_with("[!") {
        if let Some(end) = first.find(']') {
            let marker = &first[2..end];
            if let Some(kind) = CalloutKind::parse(marker) {
                let custom_title = first[end + 1..].trim();
                let title = if custom_title.is_empty() {
                    kind.label().to_string()
                } else {
                    custom_title.to_string()
                };
                let body = lines.collect::<Vec<_>>().join("\n").trim().to_string();
                return Block::Callout {
                    kind,
                    title,
                    body: RichText::plain(body),
                };
            }
        }
    }

    Block::Quote {
        content: normalize_rich_text(content),
    }
}

fn normalize_rich_text(content: RichText) -> RichText {
    let normalized = normalize_plain_text(&content.plain_text());
    if normalized.is_empty() {
        return RichText::default();
    }

    let mut out: Vec<InlineSpan> = Vec::new();
    let mut consumed = 0usize;
    let leading_trim = content
        .plain_text()
        .char_indices()
        .find(|(_, ch)| !ch.is_whitespace())
        .map(|(index, _)| index)
        .unwrap_or(0);

    for span in content.spans {
        let span_start = consumed;
        let span_end = consumed + span.text.len();
        consumed = span_end;
        if span_end <= leading_trim {
            continue;
        }

        let start = leading_trim.saturating_sub(span_start);
        let text = span.text[start..].replace('\n', " ");
        let text = text.split_whitespace().collect::<Vec<_>>().join(" ");
        if text.is_empty() {
            continue;
        }

        if let Some(last) = out.last_mut() {
            if last.bold == span.bold
                && last.italic == span.italic
                && last.code == span.code
                && last.link == span.link
            {
                if !last.text.ends_with(' ') {
                    last.text.push(' ');
                }
                last.text.push_str(&text);
                continue;
            }
        }

        out.push(InlineSpan {
            text,
            bold: span.bold,
            italic: span.italic,
            code: span.code,
            link: span.link,
        });
    }

    RichText { spans: out }
}
