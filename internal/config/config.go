package config

type Theme struct {
	Name            string
	DisplayName     string
	Background      string
	Heading         string
	Body            string
	Muted           string
	Border          string
	Code            string
	CodeBackground  string
	InlineCode      string
	Link            string
	TableHeader     string
	Quote           string
	CalloutNote     string
	CalloutTip      string
	CalloutWarning  string
	MatchHighlight  string
	BlockGap        int
	CodeLineNumbers bool
}

func ThemeNames() []string {
	return []string{"tomorrow-night", "github", "mono", "catppuccin-latte", "catppuccin-mocha", "claude", "everforest", "jellybeans", "gotham"}
}

func ThemeByName(name string) Theme {
	switch name {
	case "github", "daylight":
		return Theme{
			Name:            "github",
			DisplayName:     "GitHub Light",
			Background:      "#ffffff",
			Heading:         "#1f2328",
			Body:            "#1f2328",
			Muted:           "#656d76",
			Border:          "#d0d7de",
			Code:            "#1f2328",
			CodeBackground:  "#f6f8fa",
			InlineCode:      "#8250df",
			Link:            "#0969da",
			TableHeader:     "#1f2328",
			Quote:           "#656d76",
			CalloutNote:     "#0969da",
			CalloutTip:      "#1a7f37",
			CalloutWarning:  "#9a6700",
			MatchHighlight:  "#ffe8a3",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "catppuccin-latte", "latte", "cappuccino":
		return Theme{
			Name:            "catppuccin-latte",
			DisplayName:     "Catppuccin Latte",
			Background:      "#eff1f5",
			Heading:         "#8839ef",
			Body:            "#4c4f69",
			Muted:           "#6c6f85",
			Border:          "#acb0be",
			Code:            "#40a02b",
			CodeBackground:  "#e6e9ef",
			InlineCode:      "#fe640b",
			Link:            "#1e66f5",
			TableHeader:     "#4c4f69",
			Quote:           "#7287fd",
			CalloutNote:     "#1e66f5",
			CalloutTip:      "#40a02b",
			CalloutWarning:  "#df8e1d",
			MatchHighlight:  "#bcc0cc",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "catppuccin-mocha", "mocha":
		return Theme{
			Name:            "catppuccin-mocha",
			DisplayName:     "Catppuccin Mocha",
			Background:      "#1e1e2e",
			Heading:         "#cba6f7",
			Body:            "#cdd6f4",
			Muted:           "#a6adc8",
			Border:          "#6c7086",
			Code:            "#a6e3a1",
			CodeBackground:  "#181825",
			InlineCode:      "#fab387",
			Link:            "#89b4fa",
			TableHeader:     "#cdd6f4",
			Quote:           "#b4befe",
			CalloutNote:     "#89b4fa",
			CalloutTip:      "#a6e3a1",
			CalloutWarning:  "#f9e2af",
			MatchHighlight:  "#45475a",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "claude":
		return Theme{
			Name:            "claude",
			DisplayName:     "Claude",
			Background:      "#faf9f5",
			Heading:         "#b85c38",
			Body:            "#191919",
			Muted:           "#6b6259",
			Border:          "#d8d0c7",
			Code:            "#2a211c",
			CodeBackground:  "#f1ebe3",
			InlineCode:      "#c65f3a",
			Link:            "#b85c38",
			TableHeader:     "#191919",
			Quote:           "#7a685a",
			CalloutNote:     "#b85c38",
			CalloutTip:      "#4f7d45",
			CalloutWarning:  "#9f5d00",
			MatchHighlight:  "#ead7c4",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "everforest":
		return Theme{
			Name:            "everforest",
			DisplayName:     "Everforest Dark",
			Background:      "#2d353b",
			Heading:         "#7fbbb3",
			Body:            "#d3c6aa",
			Muted:           "#859289",
			Border:          "#475258",
			Code:            "#a7c080",
			CodeBackground:  "#232a2e",
			InlineCode:      "#e69875",
			Link:            "#83c092",
			TableHeader:     "#d3c6aa",
			Quote:           "#d699b6",
			CalloutNote:     "#7fbbb3",
			CalloutTip:      "#a7c080",
			CalloutWarning:  "#dbbc7f",
			MatchHighlight:  "#dbbc7f",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "jellybeans":
		return Theme{
			Name:            "jellybeans",
			DisplayName:     "Jellybeans",
			Background:      "#151515",
			Heading:         "#8197bf",
			Body:            "#e8e8d3",
			Muted:           "#888888",
			Border:          "#404040",
			Code:            "#99ad6a",
			CodeBackground:  "#1c1c1c",
			InlineCode:      "#ffb964",
			Link:            "#8fbfdc",
			TableHeader:     "#e8e8d3",
			Quote:           "#c6b6ee",
			CalloutNote:     "#8fbfdc",
			CalloutTip:      "#99ad6a",
			CalloutWarning:  "#fad07a",
			MatchHighlight:  "#fad07a",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "gotham":
		return Theme{
			Name:            "gotham",
			DisplayName:     "Gotham",
			Background:      "#0c1014",
			Heading:         "#599cab",
			Body:            "#99d1ce",
			Muted:           "#245361",
			Border:          "#0a3749",
			Code:            "#2aa889",
			CodeBackground:  "#11151c",
			InlineCode:      "#d26937",
			Link:            "#33859e",
			TableHeader:     "#d3ebe9",
			Quote:           "#888ca6",
			CalloutNote:     "#33859e",
			CalloutTip:      "#2aa889",
			CalloutWarning:  "#edb443",
			MatchHighlight:  "#edb443",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "mono":
		t := ThemeByName("tomorrow-night")
		t.Name = "mono"
		t.DisplayName = "Mono"
		t.Background = "#111111"
		t.Heading = "#f5f5f5"
		t.Body = "#f5f5f5"
		t.Muted = "#a3a3a3"
		t.Border = "#333333"
		t.Code = "#f5f5f5"
		t.CodeBackground = "#2b2b2b"
		t.InlineCode = "#f5f5f5"
		t.Link = "#c084fc"
		t.TableHeader = "#f5f5f5"
		t.Quote = "#a3a3a3"
		t.CalloutNote = "#f5f5f5"
		t.CalloutTip = "#f5f5f5"
		t.CalloutWarning = "#f5f5f5"
		t.MatchHighlight = "#4a4a4a"
		t.CodeLineNumbers = false
		return t
	case "tomorrow-night", "midnight":
		return Theme{
			Name:            "tomorrow-night",
			DisplayName:     "Tomorrow Night",
			Background:      "#1d1f21",
			Heading:         "#81a2be",
			Body:            "#c5c8c6",
			Muted:           "#969896",
			Border:          "#373b41",
			Code:            "#b5bd68",
			CodeBackground:  "#282a2e",
			InlineCode:      "#de935f",
			Link:            "#8abeb7",
			TableHeader:     "#c5c8c6",
			Quote:           "#b294bb",
			CalloutNote:     "#81a2be",
			CalloutTip:      "#b5bd68",
			CalloutWarning:  "#f0c674",
			MatchHighlight:  "#de935f",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	default:
		return ThemeByName("tomorrow-night")
	}
}
