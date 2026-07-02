package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandExposesExpectedFlagsAndCompletions(t *testing.T) {
	cmd := rootCommand()
	for _, name := range []string{"print", "plain", "outline", "check", "watch", "theme", "width", "no-alt-screen", "show-keys", "debug-keys"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Fatalf("missing flag %s", name)
		}
	}
	if _, _, err := cmd.Find([]string{"completions", "zsh"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cmd.Find([]string{"version"}); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultThemeIsMono(t *testing.T) {
	cmd := rootCommand()
	flag := cmd.Flags().Lookup("theme")
	if flag == nil {
		t.Fatal("missing theme flag")
	}
	if flag.DefValue != "mono" {
		t.Fatalf("theme default = %q, want mono", flag.DefValue)
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := rootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"version"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != "0.1.0" {
		t.Fatalf("version output = %q", got)
	}
}
