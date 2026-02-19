Here is a paragraph before a code block that should be rewrapped because it is too long for sixty columns.

```go
func main() {
    // This is a very long comment inside a code block that must not be rewrapped regardless of length
    fmt.Println("Hello, world! This is a long string that should stay on one line.")
}
```

Paragraph between code blocks that is also quite long and needs to be rewrapped to fit the column.

```
plain code block without language
    this should be preserved exactly as-is
        including indentation
```

    indented code block line one
    indented code block line two - this is very long and should not be rewrapped at all

Final paragraph after all the code blocks that should be rewrapped properly.
