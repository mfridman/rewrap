package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandGlobs(t *testing.T) {
	t.Parallel()

	// Setup a temp directory tree:
	//
	// 	root/
	// 	  a.go
	// 	  b.txt
	// 	  sub/
	// 	    c.go
	// 	    d.txt
	// 	    deep/
	// 	      e.go
	// 	  empty/
	setup := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		dirs := []string{
			filepath.Join(root, "sub", "deep"),
			filepath.Join(root, "empty"),
		}
		for _, d := range dirs {
			require.NoError(t, os.MkdirAll(d, 0o755))
		}
		files := []string{
			filepath.Join(root, "a.go"),
			filepath.Join(root, "b.txt"),
			filepath.Join(root, "sub", "c.go"),
			filepath.Join(root, "sub", "d.txt"),
			filepath.Join(root, "sub", "deep", "e.go"),
		}
		for _, f := range files {
			require.NoError(t, os.WriteFile(f, []byte("// "+f), 0o644))
		}
		return root
	}

	t.Run("nil_args", func(t *testing.T) {
		t.Parallel()
		got, err := expandGlobs(nil, nil)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("empty_args", func(t *testing.T) {
		t.Parallel()
		got, err := expandGlobs([]string{}, nil)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("literal_paths_passed_through", func(t *testing.T) {
		t.Parallel()
		// Literal paths are kept as-is, even if they don't exist (expandGlobs doesn't validate
		// them).
		args := []string{"foo.go", "bar/baz.txt"}
		got, err := expandGlobs(args, nil)
		require.NoError(t, err)
		require.Equal(t, args, got)
	})

	t.Run("single_level_glob", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		got, err := expandGlobs([]string{filepath.Join(root, "*.go")}, nil)
		require.NoError(t, err)
		require.Equal(t, []string{filepath.Join(root, "a.go")}, got)
	})

	t.Run("single_level_glob_multiple_matches", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		got, err := expandGlobs([]string{filepath.Join(root, "sub", "*")}, nil)
		require.NoError(t, err)
		// Should match c.go and d.txt but not the "deep" directory.
		want := []string{
			filepath.Join(root, "sub", "c.go"),
			filepath.Join(root, "sub", "d.txt"),
		}
		require.ElementsMatch(t, want, got)
	})

	t.Run("single_level_glob_filters_directories", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		// Pattern matches everything in root including "sub" and "empty" dirs.
		got, err := expandGlobs([]string{filepath.Join(root, "*")}, nil)
		require.NoError(t, err)
		for _, f := range got {
			info, err := os.Stat(f)
			require.NoError(t, err)
			require.False(t, info.IsDir(), "directories should be filtered out, got %s", f)
		}
		want := []string{
			filepath.Join(root, "a.go"),
			filepath.Join(root, "b.txt"),
		}
		require.ElementsMatch(t, want, got)
	})

	t.Run("recursive_glob_go_files", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		got, err := expandGlobs([]string{root + string(filepath.Separator) + "**/*.go"}, nil)
		require.NoError(t, err)
		want := []string{
			filepath.Join(root, "a.go"),
			filepath.Join(root, "sub", "c.go"),
			filepath.Join(root, "sub", "deep", "e.go"),
		}
		require.ElementsMatch(t, want, got)
	})

	t.Run("recursive_glob_all_files", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		got, err := expandGlobs([]string{root + string(filepath.Separator) + "**/*"}, nil)
		require.NoError(t, err)
		want := []string{
			filepath.Join(root, "a.go"),
			filepath.Join(root, "b.txt"),
			filepath.Join(root, "sub", "c.go"),
			filepath.Join(root, "sub", "d.txt"),
			filepath.Join(root, "sub", "deep", "e.go"),
		}
		require.ElementsMatch(t, want, got)
	})

	t.Run("recursive_glob_no_prefix", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		// Change to the temp dir so "**/*.go" walks from ".".
		orig, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(root))
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(orig))
		})

		got, err := expandGlobs([]string{"**/*.go"}, nil)
		require.NoError(t, err)
		want := []string{
			"a.go",
			filepath.Join("sub", "c.go"),
			filepath.Join("sub", "deep", "e.go"),
		}
		require.ElementsMatch(t, want, got)
	})

	t.Run("question_mark_glob", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		got, err := expandGlobs([]string{filepath.Join(root, "?.go")}, nil)
		require.NoError(t, err)
		require.Equal(t, []string{filepath.Join(root, "a.go")}, got)
	})

	t.Run("bracket_glob", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		got, err := expandGlobs([]string{filepath.Join(root, "[ab].*")}, nil)
		require.NoError(t, err)
		want := []string{
			filepath.Join(root, "a.go"),
			filepath.Join(root, "b.txt"),
		}
		require.ElementsMatch(t, want, got)
	})

	t.Run("no_match_single_glob_error", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		_, err := expandGlobs([]string{filepath.Join(root, "*.nonexistent")}, nil)
		require.Error(t, err)
	})

	t.Run("no_match_recursive_glob_error", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		_, err := expandGlobs([]string{root + string(filepath.Separator) + "**/*.nonexistent"}, nil)
		require.Error(t, err)
	})

	t.Run("mixed_literal_and_glob", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		literal := filepath.Join(root, "a.go")
		glob := filepath.Join(root, "sub", "*.go")
		got, err := expandGlobs([]string{literal, glob}, nil)
		require.NoError(t, err)
		want := []string{
			filepath.Join(root, "a.go"),
			filepath.Join(root, "sub", "c.go"),
		}
		require.Equal(t, want, got)
	})

	t.Run("recursive_glob_nonexistent_root", func(t *testing.T) {
		t.Parallel()
		_, err := expandGlobs([]string{"/nonexistent/path/**/*.go"}, nil)
		require.Error(t, err)
	})

	t.Run("empty_directory_recursive", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		_, err := expandGlobs([]string{filepath.Join(root, "empty") + string(filepath.Separator) + "**/*"}, nil)
		require.Error(t, err)
	})

	t.Run("exclude_dir_recursive", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		got, err := expandGlobs(
			[]string{root + string(filepath.Separator) + "**/*.go"},
			[]string{"deep"},
		)
		require.NoError(t, err)
		want := []string{
			filepath.Join(root, "a.go"),
			filepath.Join(root, "sub", "c.go"),
		}
		require.ElementsMatch(t, want, got)
	})

	t.Run("exclude_multiple_dirs_recursive", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		got, err := expandGlobs(
			[]string{root + string(filepath.Separator) + "**/*"},
			[]string{"sub", "empty"},
		)
		require.NoError(t, err)
		// Only root-level files remain since "sub" (and its children) are excluded.
		want := []string{
			filepath.Join(root, "a.go"),
			filepath.Join(root, "b.txt"),
		}
		require.ElementsMatch(t, want, got)
	})

	t.Run("exclude_dir_single_level_glob", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		// Single-level glob in sub/ -- excluding "deep" shouldn't affect direct matches.
		got, err := expandGlobs(
			[]string{filepath.Join(root, "sub", "*")},
			[]string{"deep"},
		)
		require.NoError(t, err)
		want := []string{
			filepath.Join(root, "sub", "c.go"),
			filepath.Join(root, "sub", "d.txt"),
		}
		require.ElementsMatch(t, want, got)
	})

	t.Run("exclude_parent_dir_skips_children", func(t *testing.T) {
		t.Parallel()
		root := setup(t)
		// Excluding "sub" should also skip "sub/deep/e.go".
		got, err := expandGlobs(
			[]string{root + string(filepath.Separator) + "**/*.go"},
			[]string{"sub"},
		)
		require.NoError(t, err)
		require.Equal(t, []string{filepath.Join(root, "a.go")}, got)
	})
}
