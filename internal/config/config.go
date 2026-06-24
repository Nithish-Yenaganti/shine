package config

type Theme struct {
	Name            string
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
	return []string{"midnight", "daylight", "mono", "catppuccin-latte", "catppuccin-mocha", "claude"}
}

func ThemeByName(name string) Theme {
	switch name {
	case "daylight":
		return Theme{
			Name:            "daylight",
			Background:      "#ffffff",
			Heading:         "#0f4c81",
			Body:            "#24292f",
			Muted:           "#57606a",
			Border:          "#8c959f",
			Code:            "#24292f",
			CodeBackground:  "#f6f8fa",
			InlineCode:      "#8250df",
			Link:            "#0969da",
			TableHeader:     "#24292f",
			Quote:           "#6e7781",
			CalloutNote:     "#0969da",
			CalloutTip:      "#1a7f37",
			CalloutWarning:  "#9a6700",
			MatchHighlight:  "#ffe8a3",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "catppuccin-latte", "cappuccino":
		return Theme{
			Name:            "catppuccin-latte",
			Background:      "#eff1f5",
			Heading:         "#8839ef",
			Body:            "#4c4f69",
			Muted:           "#6c6f85",
			Border:          "#9ca0b0",
			Code:            "#40a02b",
			CodeBackground:  "#e6e9ef",
			InlineCode:      "#fe640b",
			Link:            "#1e66f5",
			TableHeader:     "#4c4f69",
			Quote:           "#7287fd",
			CalloutNote:     "#1e66f5",
			CalloutTip:      "#40a02b",
			CalloutWarning:  "#df8e1d",
			MatchHighlight:  "#ccd0da",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "catppuccin-mocha", "mocha":
		return Theme{
			Name:            "catppuccin-mocha",
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
	case "mono":
		t := ThemeByName("midnight")
		t.Name = "mono"
		t.Background = ""
		t.Heading = "15"
		t.Body = "15"
		t.Muted = "8"
		t.Border = "8"
		t.Code = "15"
		t.CodeBackground = "0"
		t.InlineCode = "15"
		t.Link = "14"
		t.TableHeader = "15"
		t.Quote = "8"
		t.CalloutNote = "15"
		t.CalloutTip = "15"
		t.CalloutWarning = "15"
		t.CodeLineNumbers = false
		return t
	default:
		return Theme{
			Name:            "midnight",
			Heading:         "#8bd5ff",
			Body:            "#e7e7e7",
			Muted:           "#8a8f98",
			Border:          "#3a3f4b",
			Code:            "#b7f7c8",
			CodeBackground:  "#151922",
			InlineCode:      "#ffb86c",
			Link:            "#7dcfff",
			TableHeader:     "#ffffff",
			Quote:           "#c6a0ff",
			CalloutNote:     "#8bd5ff",
			CalloutTip:      "#7ee787",
			CalloutWarning:  "#ffd166",
			MatchHighlight:  "#ffd166",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	}
}
