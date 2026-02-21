//go:build linux && amd64

package example

//go:generate stringer -type=Kind

//go:embed templates/*.html
var templates string

// This is a normal doc comment that should be rewrapped to fit within the column width properly.

//go:noinline
func Example() {}

//nolint:errcheck
func Another() {}

// Regular comment before a directive.
//
//go:nosplit
func Third() {}
