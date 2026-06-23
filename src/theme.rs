use ratatui::style::{Color, Modifier, Style};
use serde::Deserialize;

#[derive(Debug, Clone, Deserialize)]
#[serde(default)]
pub struct Theme {
    pub name: String,
    pub heading: String,
    pub body: String,
    pub muted: String,
    pub border: String,
    pub code: String,
    pub inline_code: String,
    pub link: String,
    pub table_header: String,
    pub code_line_number: String,
    pub quote: String,
    pub callout_note: String,
    pub callout_tip: String,
    pub callout_warning: String,
    pub selection: String,
    pub match_highlight: String,
    pub block_padding: Option<u16>,
    pub code_line_numbers: Option<bool>,
}

impl Default for Theme {
    fn default() -> Self {
        Self {
            name: "midnight".to_string(),
            heading: "#8bd5ff".to_string(),
            body: "#e7e7e7".to_string(),
            muted: "#8a8f98".to_string(),
            border: "#3a3f4b".to_string(),
            code: "#b7f7c8".to_string(),
            inline_code: "#ffb86c".to_string(),
            link: "#7dcfff".to_string(),
            table_header: "#ffffff".to_string(),
            code_line_number: "#6c7280".to_string(),
            quote: "#c6a0ff".to_string(),
            callout_note: "#8bd5ff".to_string(),
            callout_tip: "#7ee787".to_string(),
            callout_warning: "#ffd166".to_string(),
            selection: "#26364d".to_string(),
            match_highlight: "#ffd166".to_string(),
            block_padding: Some(1),
            code_line_numbers: Some(true),
        }
    }
}

impl Theme {
    pub fn light() -> Self {
        Self {
            name: "daylight".to_string(),
            heading: "#0b5cad".to_string(),
            body: "#1f2328".to_string(),
            muted: "#59636e".to_string(),
            border: "#d0d7de".to_string(),
            code: "#116329".to_string(),
            inline_code: "#953800".to_string(),
            link: "#0969da".to_string(),
            table_header: "#24292f".to_string(),
            code_line_number: "#8c959f".to_string(),
            quote: "#6639ba".to_string(),
            callout_note: "#0969da".to_string(),
            callout_tip: "#1a7f37".to_string(),
            callout_warning: "#9a6700".to_string(),
            selection: "#ddf4ff".to_string(),
            match_highlight: "#fff8c5".to_string(),
            block_padding: Some(1),
            code_line_numbers: Some(true),
        }
    }

    pub fn mono() -> Self {
        Self {
            name: "mono".to_string(),
            heading: "white".to_string(),
            body: "white".to_string(),
            muted: "gray".to_string(),
            border: "gray".to_string(),
            code: "white".to_string(),
            inline_code: "white".to_string(),
            link: "cyan".to_string(),
            table_header: "white".to_string(),
            code_line_number: "gray".to_string(),
            quote: "gray".to_string(),
            callout_note: "white".to_string(),
            callout_tip: "white".to_string(),
            callout_warning: "white".to_string(),
            selection: "gray".to_string(),
            match_highlight: "yellow".to_string(),
            block_padding: Some(1),
            code_line_numbers: Some(false),
        }
    }

    pub fn heading_style(&self) -> Style {
        Style::default()
            .fg(parse_color(&self.heading))
            .add_modifier(Modifier::BOLD)
    }

    pub fn body_style(&self) -> Style {
        Style::default().fg(parse_color(&self.body))
    }

    pub fn muted_style(&self) -> Style {
        Style::default().fg(parse_color(&self.muted))
    }

    pub fn border_style(&self) -> Style {
        Style::default().fg(parse_color(&self.border))
    }

    pub fn code_style(&self) -> Style {
        Style::default().fg(parse_color(&self.code))
    }

    pub fn inline_code_style(&self) -> Style {
        Style::default()
            .fg(parse_color(&self.inline_code))
            .add_modifier(Modifier::BOLD)
    }

    pub fn link_style(&self) -> Style {
        Style::default()
            .fg(parse_color(&self.link))
            .add_modifier(Modifier::UNDERLINED)
    }

    pub fn quote_style(&self) -> Style {
        Style::default().fg(parse_color(&self.quote))
    }

    pub fn match_style(&self) -> Style {
        Style::default()
            .fg(Color::Black)
            .bg(parse_color(&self.match_highlight))
            .add_modifier(Modifier::BOLD)
    }
}

pub fn parse_color(value: &str) -> Color {
    let value = value.trim();
    if let Some(hex) = value.strip_prefix('#') {
        if hex.len() == 6 {
            let r = u8::from_str_radix(&hex[0..2], 16);
            let g = u8::from_str_radix(&hex[2..4], 16);
            let b = u8::from_str_radix(&hex[4..6], 16);
            if let (Ok(r), Ok(g), Ok(b)) = (r, g, b) {
                return Color::Rgb(r, g, b);
            }
        }
    }

    match value.to_ascii_lowercase().as_str() {
        "black" => Color::Black,
        "red" => Color::Red,
        "green" => Color::Green,
        "yellow" => Color::Yellow,
        "blue" => Color::Blue,
        "magenta" => Color::Magenta,
        "cyan" => Color::Cyan,
        "gray" | "grey" => Color::Gray,
        "white" => Color::White,
        _ => Color::White,
    }
}
