package cli

import "testing"

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
}
