# rewrap

<img align="right" width="125" src=".github/rewrap_logo.png">

CLI tool that rewraps comment blocks and text to a specified column width. Language is detected from
the file extension, or can be overridden with `--lang`.

Non-comment code is passed through unchanged. Only comment blocks are reflowed.

## Install

```
go install github.com/mfridman/rewrap@latest
```

## Usage

```
rewrap [flags] [files...]
```

Flags:

- `-c`, `--column` - wrapping column width (default 100)
- `-v`, `--verbose` - print each file path when writing
- `-w`, `--write` - write result to file instead of stdout
- `--tab-width` - tab display width for column calculations (default 4)
- `--lang` - override language detection (e.g., `go`, `python`, `markdown`, `text`)
- `--exclude` - comma-separated directory names to exclude (e.g., `testdata,vendor`)

## Examples

Rewrap a Go file to 80 columns and print to stdout:

```
rewrap -c 80 main.go
```

Rewrap in place:

```
rewrap -c 100 -w main.go
```

Glob patterns (quote to prevent shell expansion):

```
rewrap -c 100 -w 'wrap/*.go'
rewrap -c 100 -w '**/*.go'
rewrap -c 100 -w '**/*.go' --exclude testdata,vendor
```

Pipe through stdin:

```
cat main.go | rewrap --lang go -c 80
```

## Supported languages

Go, C, C++, Java, JavaScript, TypeScript, Python, Shell, Ruby, Rust, Markdown.

Use `--lang text` to treat input as plain text (rewraps everything).

## Language-specific behavior

- **Go** - uses `go/doc/comment` for rewrapping, so doc comment syntax (headings, lists, code
  blocks, links) is handled correctly.
- **Markdown** - uses AST-based parsing. Only paragraph text is rewrapped; headings, code blocks,
  lists, blockquotes, tables, and other structural elements are preserved verbatim.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
