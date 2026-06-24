package config

type Theme struct {
	Name            string
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

func ThemeByName(name string) Theme {
	switch name {
	case "daylight":
		return Theme{
			Name:            "daylight",
			Heading:         "#0b5cad",
			Body:            "#1f2328",
			Muted:           "#667085",
			Border:          "#d0d7de",
			Code:            "#116329",
			CodeBackground:  "#f6f8fa",
			InlineCode:      "#953800",
			Link:            "#0969da",
			TableHeader:     "#24292f",
			Quote:           "#6639ba",
			CalloutNote:     "#0969da",
			CalloutTip:      "#1a7f37",
			CalloutWarning:  "#9a6700",
			MatchHighlight:  "#fff8c5",
			BlockGap:        1,
			CodeLineNumbers: true,
		}
	case "mono":
		t := ThemeByName("midnight")
		t.Name = "mono"
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
