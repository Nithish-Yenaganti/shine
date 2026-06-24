package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"shine/internal/config"
	"shine/internal/inspect"
	"shine/internal/parser"
	"shine/internal/render"
	"shine/internal/source"
	shinetui "shine/internal/tui"
	"shine/internal/version"
)

type options struct {
	watch       bool
	print       bool
	plain       bool
	outline     bool
	check       bool
	theme       string
	width       int
	noAltScreen bool
	showKeys    bool
	debugKeys   bool
}

func Execute() {
	if err := rootCommand().Execute(); err != nil {
		var ee exitError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCommand() *cobra.Command {
	opts := &options{theme: "midnight", width: 88}
	cmd := &cobra.Command{
		Use:           "shine [file]",
		Short:         "Preview Markdown beautifully inside the terminal",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version.String(),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := source.Read(args)
			if err != nil {
				return err
			}
			theme := config.ThemeByName(opts.theme)
			if opts.print || opts.plain || opts.outline || opts.check {
				doc, err := parser.Parse([]byte(src.Content), src.Name)
				if err != nil {
					return err
				}
				if opts.outline {
					fmt.Print(inspect.Outline(doc))
					return nil
				}
				if opts.check {
					warnings := inspect.Check(doc, src.Path)
					fmt.Print(inspect.FormatWarnings(warnings))
					if len(warnings) > 0 {
						return exitError{code: 2}
					}
					return nil
				}
				r := render.New(opts.width, theme)
				out := r.Render(doc)
				if opts.plain {
					out = inspect.StripANSI(out)
				}
				fmt.Print(out)
				return nil
			}
			return shinetui.Run(shinetui.Options{
				Source:       src,
				Theme:        theme,
				Watch:        opts.watch,
				UseAltScreen: !opts.noAltScreen,
				ShowKeys:     opts.showKeys,
				DebugKeys:    opts.debugKeys,
			})
		},
	}
	cmd.Flags().BoolVarP(&opts.watch, "watch", "w", false, "re-render the file when it changes")
	cmd.Flags().BoolVar(&opts.print, "print", false, "print the rendered document and exit")
	cmd.Flags().BoolVar(&opts.plain, "plain", false, "print the rendered document without ANSI styling and exit")
	cmd.Flags().BoolVar(&opts.outline, "outline", false, "print a Markdown heading outline and exit")
	cmd.Flags().BoolVar(&opts.check, "check", false, "check Markdown quality and exit with code 2 when warnings are found")
	cmd.Flags().StringVar(&opts.theme, "theme", "midnight", "theme preset: midnight, daylight, mono, catppuccin-latte, catppuccin-mocha, claude, everforest, jellybeans, gotham")
	cmd.Flags().IntVar(&opts.width, "width", 88, "render width for --print mode")
	cmd.Flags().BoolVar(&opts.noAltScreen, "no-alt-screen", false, "disable alternate screen mode")
	cmd.Flags().BoolVar(&opts.showKeys, "show-keys", false, "open the keyboard help panel on launch")
	cmd.Flags().BoolVar(&opts.debugKeys, "debug-keys", false, "show the last received key in the status line")
	cmd.AddCommand(completionsCommand(cmd))
	cmd.AddCommand(versionCommand())
	return cmd
}

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func completionsCommand(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completions [bash|zsh|fish|powershell]",
		Short: "Generate shell completions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return root.GenPowerShellCompletion(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell %q", args[0])
			}
		},
	}
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print shine version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version.String())
		},
	}
}
