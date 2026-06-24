package source

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Read(args []string) (Source, error) {
	if len(args) > 0 {
		path := filepath.Clean(args[0])
		content, err := os.ReadFile(path)
		if err != nil {
			return Source{}, err
		}
		return Source{Name: path, Path: path, Content: string(content)}, nil
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return Source{}, err
	}
	if stat.Mode()&os.ModeCharDevice != 0 {
		return Source{}, fmt.Errorf("provide a Markdown file or pipe Markdown into stdin")
	}

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return Source{}, err
	}
	return Source{Name: "stdin", Content: string(content)}, nil
}
