package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"shine/internal/config"
	"shine/internal/parser"
	"shine/internal/render"
	"shine/internal/source"
	shinetui "shine/internal/tui"
)

type options struct {
	watch       bool
	print       bool
	theme       string
	width       int
	noAltScreen bool
}

func Execute() {
	if err := rootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCommand() *cobra.Command {
	opts := &options{theme: "midnight", width: 88}
	cmd := &cobra.Command{
		Use:   "shine [file]",
		Short: "Preview Markdown beautifully inside the terminal",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := source.Read(args)
			if err != nil {
				return err
			}
			theme := config.ThemeByName(opts.theme)
			if opts.print {
				doc, err := parser.Parse([]byte(src.Content), src.Name)
				if err != nil {
					return err
				}
				r := render.New(opts.width, theme)
				fmt.Print(r.Render(doc))
				return nil
			}
			return shinetui.Run(shinetui.Options{
				Source:       src,
				Theme:        theme,
				Watch:        opts.watch,
				UseAltScreen: !opts.noAltScreen,
			})
		},
	}
	cmd.Flags().BoolVarP(&opts.watch, "watch", "w", false, "re-render the file when it changes")
	cmd.Flags().BoolVar(&opts.print, "print", false, "print the rendered document and exit")
	cmd.Flags().StringVar(&opts.theme, "theme", "midnight", "theme preset: midnight, daylight, mono")
	cmd.Flags().IntVar(&opts.width, "width", 88, "render width for --print mode")
	cmd.Flags().BoolVar(&opts.noAltScreen, "no-alt-screen", false, "disable alternate screen mode")
	cmd.AddCommand(completionsCommand(cmd))
	return cmd
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
