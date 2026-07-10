package config

import "testing"

func TestThemeNamesResolve(t *testing.T) {
	for _, name := range ThemeNames() {
		theme := ThemeByName(name)
		if theme.Name != name {
			t.Fatalf("expected theme %q to resolve by name, got %q", name, theme.Name)
		}
		if theme.Body == "" || theme.Heading == "" || theme.Border == "" {
			t.Fatalf("theme %q is missing required colors: %+v", name, theme)
		}
		if theme.DisplayName == "" {
			t.Fatalf("theme %q is missing a display name", name)
		}
	}
}

func TestThemeAliasesResolveToRealThemeNames(t *testing.T) {
	tests := map[string]string{
		"cappuccino": "catppuccin-latte",
		"daylight":   "github",
		"latte":      "catppuccin-latte",
		"midnight":   "tomorrow-night",
		"mocha":      "catppuccin-mocha",
	}
	for alias, want := range tests {
		if got := ThemeByName(alias).Name; got != want {
			t.Fatalf("expected alias %q to resolve to %q, got %q", alias, want, got)
		}
	}
}

func TestGitHubLightUsesPrimerPalette(t *testing.T) {
	theme := ThemeByName("github")
	if theme.DisplayName != "GitHub Light" {
		t.Fatalf("expected GitHub display name, got %q", theme.DisplayName)
	}
	if theme.Background != "#ffffff" || theme.Body != "#1f2328" || theme.Heading != "#1f2328" {
		t.Fatalf("unexpected GitHub base colors: background=%q body=%q heading=%q", theme.Background, theme.Body, theme.Heading)
	}
	if theme.Muted != "#656d76" || theme.Border != "#d0d7de" || theme.Link != "#0969da" {
		t.Fatalf("unexpected GitHub support colors: muted=%q border=%q link=%q", theme.Muted, theme.Border, theme.Link)
	}
}

func TestTomorrowNightThemeUsesTomorrowNightPalette(t *testing.T) {
	theme := ThemeByName("tomorrow-night")
	if theme.Background != "#1d1f21" {
		t.Fatalf("expected tomorrow-night background to use Tomorrow Night base, got %q", theme.Background)
	}
	if theme.Body != "#c5c8c6" || theme.Muted != "#969896" || theme.Border != "#373b41" {
		t.Fatalf("expected tomorrow-night neutrals to use Tomorrow Night palette, got body=%q muted=%q border=%q", theme.Body, theme.Muted, theme.Border)
	}
	if theme.Heading != "#81a2be" || theme.Link != "#8abeb7" || theme.InlineCode != "#de935f" {
		t.Fatalf("expected tomorrow-night accents to use Tomorrow Night palette, got heading=%q link=%q inline=%q", theme.Heading, theme.Link, theme.InlineCode)
	}
}

func TestClaudeThemeUsesAnthropicInspiredPalette(t *testing.T) {
	theme := ThemeByName("claude")
	if theme.Background != "#faf9f5" {
		t.Fatalf("expected Claude background to use warm ivory, got %q", theme.Background)
	}
	if theme.Body != "#191919" {
		t.Fatalf("expected Claude body to use near-black ink, got %q", theme.Body)
	}
	if theme.Heading != "#b85c38" || theme.Link != "#b85c38" || theme.CalloutNote != "#b85c38" {
		t.Fatalf("expected Claude accent colors to use Anthropic-like clay, got heading=%q link=%q note=%q", theme.Heading, theme.Link, theme.CalloutNote)
	}
}

func TestMonoThemeUsesBlackAndWhitePalette(t *testing.T) {
	theme := ThemeByName("mono")
	if theme.Background != "#111111" || theme.Body != "#f5f5f5" || theme.Heading != "#f5f5f5" {
		t.Fatalf("unexpected mono base colors: background=%q body=%q heading=%q", theme.Background, theme.Body, theme.Heading)
	}
	if theme.CodeBackground != "#2b2b2b" || theme.Code != "#f5f5f5" || theme.InlineCode != "#f5f5f5" {
		t.Fatalf("unexpected mono code colors: background=%q code=%q inline=%q", theme.CodeBackground, theme.Code, theme.InlineCode)
	}
	if theme.CodeLineNumbers {
		t.Fatalf("mono should keep code line numbers disabled")
	}
}

func TestCatppuccinLatteUsesOfficialPaletteNamesAndColors(t *testing.T) {
	theme := ThemeByName("latte")
	if theme.Name != "catppuccin-latte" || theme.DisplayName != "Catppuccin Latte" {
		t.Fatalf("latte alias resolved to name=%q display=%q", theme.Name, theme.DisplayName)
	}
	if theme.Background != "#eff1f5" || theme.Body != "#4c4f69" || theme.Heading != "#8839ef" {
		t.Fatalf("unexpected Catppuccin Latte base colors: background=%q body=%q heading=%q", theme.Background, theme.Body, theme.Heading)
	}
	if theme.Border != "#acb0be" || theme.MatchHighlight != "#bcc0cc" {
		t.Fatalf("unexpected Catppuccin Latte surface colors: border=%q match=%q", theme.Border, theme.MatchHighlight)
	}
}

func TestAdditionalDarkThemePalettes(t *testing.T) {
	tests := map[string]struct {
		background string
		body       string
		heading    string
	}{
		"everforest": {"#2d353b", "#d3c6aa", "#7fbbb3"},
		"jellybeans": {"#151515", "#e8e8d3", "#8197bf"},
		"gotham":     {"#0c1014", "#99d1ce", "#599cab"},
	}
	for name, want := range tests {
		theme := ThemeByName(name)
		if theme.Background != want.background || theme.Body != want.body || theme.Heading != want.heading {
			t.Fatalf("unexpected %s palette: background=%q body=%q heading=%q", name, theme.Background, theme.Body, theme.Heading)
		}
	}
}
