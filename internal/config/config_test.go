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
	}
}

func TestThemeAliasesResolveToRealThemeNames(t *testing.T) {
	tests := map[string]string{
		"cappuccino": "catppuccin-latte",
		"mocha":      "catppuccin-mocha",
	}
	for alias, want := range tests {
		if got := ThemeByName(alias).Name; got != want {
			t.Fatalf("expected alias %q to resolve to %q, got %q", alias, want, got)
		}
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
