package example

// Package example demonstrates various Go doc comment features that should be handled correctly by the rewrapper.
//
// This is a paragraph that follows the package declaration and should be rewrapped to fit within the column width.
//
// # Heading
//
// This paragraph comes after a heading. It should be rewrapped but the heading itself should be preserved.
//
// Here is a list:
//   - First item that is long enough to potentially need rewrapping within the list context.
//   - Second item.
//   - Third item with some extra text.
//
// And a code block:
//
//	func hello() {
//	    fmt.Println("hello")
//	}
//
// The code block above should not be rewrapped.
func Example() {}
