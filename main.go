package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mfridman/rewrap/wrap"
	"github.com/pressly/cli"
)

func main() {
	root := &cli.Command{
		Name:      "rewrap",
		Usage:     "rewrap [flags] [files...]",
		ShortHelp: "Rewrap comment blocks and text to a specified column width",
		Flags: cli.FlagsFunc(func(f *flag.FlagSet) {
			f.Int("column", 100, "wrapping column width")
			f.Bool("write", false, "write result to file instead of stdout")
			f.Int("tab-width", 4, "tab display width for column calculations")
			f.String("lang", "", "override language detection")
			f.Bool("verbose", false, "print each file path when writing")
			f.String("exclude", "", "comma-separated directory names to exclude")
		}),
		FlagOptions: []cli.FlagOption{
			{Name: "column", Short: "c"},
			{Name: "write", Short: "w"},
			{Name: "verbose", Short: "v"},
		},
		UsageFunc: func(c *cli.Command) string {
			// Temporarily clear UsageFunc to avoid infinite recursion with DefaultUsage.
			c.UsageFunc = nil
			s := cli.DefaultUsage(c)
			return s + "\n\n" + `Examples:
  rewrap -c 80 main.go                                  Rewrap a single file
  rewrap -c 100 -w main.go                              Rewrap and write in place
  rewrap -c 100 'wrap/*.go'                             Glob: all Go files in wrap/
  rewrap -c 100 '**/*.go'                               Recursive glob: all Go files
  rewrap -c 100 -w '**/*.go' --exclude testdata,vendor  Skip directories
  cat main.go | rewrap --lang go                        Pipe through stdin`
		},
		Exec: execRoot,
	}
	if err := cli.ParseAndRun(context.Background(), root, os.Args[1:], nil); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func execRoot(ctx context.Context, s *cli.State) error {
	column := cli.GetFlag[int](s, "column")
	write := cli.GetFlag[bool](s, "write")
	verbose := cli.GetFlag[bool](s, "verbose")
	tabWidth := cli.GetFlag[int](s, "tab-width")
	langOverride := cli.GetFlag[string](s, "lang")

	var excludeDirs []string
	if e := cli.GetFlag[string](s, "exclude"); e != "" {
		for d := range strings.SplitSeq(e, ",") {
			if d = strings.TrimSpace(d); d != "" {
				excludeDirs = append(excludeDirs, d)
			}
		}
	}

	files, err := expandGlobs(s.Args, excludeDirs)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		// Check if stdin is a pipe.
		stat, err := os.Stdin.Stat()
		if err != nil {
			return fmt.Errorf("stat stdin: %w", err)
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("usage: rewrap [flags] [files...]\n\nUse -help for more information")
		}
		src, err := io.ReadAll(s.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		lang, err := resolveLanguage("", langOverride)
		if err != nil {
			return err
		}
		result := wrap.Source(src, lang, column, tabWidth)
		_, err = s.Stdout.Write(result)
		return err
	}

	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		lang, err := resolveLanguage(file, langOverride)
		if err != nil {
			return err
		}
		result := wrap.Source(src, lang, column, tabWidth)
		if write {
			info, err := os.Stat(file)
			if err != nil {
				return fmt.Errorf("stat %s: %w", file, err)
			}
			if err := os.WriteFile(file, result, info.Mode().Perm()); err != nil {
				return fmt.Errorf("write %s: %w", file, err)
			}
			if verbose {
				_, _ = fmt.Fprintln(s.Stdout, file)
			}
		} else {
			if _, err := s.Stdout.Write(result); err != nil {
				return err
			}
		}
	}
	return nil
}

func expandGlobs(args []string, excludeDirs []string) ([]string, error) {
	var files []string
	for _, arg := range args {
		if !strings.ContainsAny(arg, "*?[") {
			files = append(files, arg)
			continue
		}
		var matches []string
		if strings.Contains(arg, "**") {
			// Handle recursive glob patterns with filepath.WalkDir.
			prefix, suffix, _ := strings.Cut(arg, "**")
			root := prefix
			if root == "" {
				root = "."
			}
			suffix = strings.TrimPrefix(suffix, string(filepath.Separator))
			if suffix == "" {
				suffix = "*"
			}
			err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					if isExcludedDir(d.Name(), excludeDirs) {
						return filepath.SkipDir
					}
					return nil
				}
				matched, matchErr := filepath.Match(suffix, filepath.Base(path))
				if matchErr != nil {
					return matchErr
				}
				if matched {
					matches = append(matches, path)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walk %s: %w", arg, err)
			}
		} else {
			var err error
			matches, err = filepath.Glob(arg)
			if err != nil {
				return nil, fmt.Errorf("glob %s: %w", arg, err)
			}
			// Filter out directories and excluded paths.
			filtered := matches[:0]
			for _, m := range matches {
				info, err := os.Stat(m)
				if err != nil {
					return nil, err
				}
				if info.IsDir() {
					continue
				}
				if containsExcludedDir(m, excludeDirs) {
					continue
				}
				filtered = append(filtered, m)
			}
			matches = filtered
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("pattern %q matched no files", arg)
		}
		files = append(files, matches...)
	}
	return files, nil
}

func isExcludedDir(name string, excludeDirs []string) bool {
	return slices.Contains(excludeDirs, name)
}

func containsExcludedDir(path string, excludeDirs []string) bool {
	for part := range strings.SplitSeq(filepath.ToSlash(path), "/") {
		if isExcludedDir(part, excludeDirs) {
			return true
		}
	}
	return false
}

func resolveLanguage(filename, langOverride string) (*wrap.Language, error) {
	if langOverride == "text" {
		return nil, nil
	}
	if langOverride != "" {
		lang := wrap.LanguageFromName(langOverride)
		if lang == nil {
			return nil, fmt.Errorf("unknown language: %s", langOverride)
		}
		return lang, nil
	}
	if filename != "" {
		return wrap.LanguageFromFilename(filename), nil
	}
	return nil, nil
}
