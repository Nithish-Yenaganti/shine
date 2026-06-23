#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Document {
    pub source_name: Option<String>,
    pub blocks: Vec<Block>,
    pub headings: Vec<HeadingRef>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct HeadingRef {
    pub level: u8,
    pub text: String,
    pub block_index: usize,
}

#[derive(Debug, Clone, PartialEq, Eq, Default)]
pub struct RichText {
    pub spans: Vec<InlineSpan>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct InlineSpan {
    pub text: String,
    pub bold: bool,
    pub italic: bool,
    pub code: bool,
    pub link: Option<String>,
}

impl InlineSpan {
    pub fn plain(text: impl Into<String>) -> Self {
        Self {
            text: text.into(),
            bold: false,
            italic: false,
            code: false,
            link: None,
        }
    }
}

impl RichText {
    pub fn plain(text: impl Into<String>) -> Self {
        let text = text.into();
        if text.is_empty() {
            Self::default()
        } else {
            Self {
                spans: vec![InlineSpan::plain(text)],
            }
        }
    }

    pub fn plain_text(&self) -> String {
        self.spans
            .iter()
            .map(|span| span.text.as_str())
            .collect::<String>()
    }

    pub fn normalized_plain_text(&self) -> String {
        normalize_plain_text(&self.plain_text())
    }

    pub fn is_empty(&self) -> bool {
        self.spans.iter().all(|span| span.text.trim().is_empty())
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Block {
    Heading {
        level: u8,
        content: RichText,
    },
    Paragraph {
        content: RichText,
    },
    CodeBlock {
        language: Option<String>,
        code: String,
    },
    Quote {
        content: RichText,
    },
    Callout {
        kind: CalloutKind,
        title: String,
        body: RichText,
    },
    List {
        ordered: bool,
        items: Vec<ListItem>,
    },
    Table {
        headers: Vec<String>,
        rows: Vec<Vec<String>>,
    },
    Divider,
    Image {
        alt: String,
        url: String,
    },
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ListItem {
    pub content: RichText,
    pub checked: Option<bool>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CalloutKind {
    Note,
    Tip,
    Warning,
    Important,
    Caution,
}

impl CalloutKind {
    pub fn label(self) -> &'static str {
        match self {
            Self::Note => "NOTE",
            Self::Tip => "TIP",
            Self::Warning => "WARNING",
            Self::Important => "IMPORTANT",
            Self::Caution => "CAUTION",
        }
    }

    pub fn parse(value: &str) -> Option<Self> {
        match value.trim().to_ascii_uppercase().as_str() {
            "NOTE" => Some(Self::Note),
            "TIP" => Some(Self::Tip),
            "WARNING" | "WARN" => Some(Self::Warning),
            "IMPORTANT" => Some(Self::Important),
            "CAUTION" | "DANGER" => Some(Self::Caution),
            _ => None,
        }
    }
}

impl Document {
    pub fn new(source_name: Option<String>, blocks: Vec<Block>) -> Self {
        let headings = blocks
            .iter()
            .enumerate()
            .filter_map(|(block_index, block)| match block {
                Block::Heading { level, content } => Some(HeadingRef {
                    level: *level,
                    text: content.normalized_plain_text(),
                    block_index,
                }),
                _ => None,
            })
            .collect();

        Self {
            source_name,
            blocks,
            headings,
        }
    }

    pub fn block_count(&self) -> usize {
        self.blocks.len()
    }
}

pub fn normalize_plain_text(value: &str) -> String {
    value
        .lines()
        .map(str::trim)
        .filter(|line| !line.is_empty())
        .collect::<Vec<_>>()
        .join(" ")
        .trim()
        .to_string()
}
