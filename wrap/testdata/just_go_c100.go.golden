package example

// RunOptions specifies options for running a command.
type RunOptions struct {
	// Stdin, Stdout, and Stderr are the standard input, output, and error streams for the command.
	// If any of these are nil, the command will use the default streams ([os.Stdin], [os.Stdout],
	// and [os.Stderr], respectively).
	//
	// If any of these are nil, the command will use the default streams (os.Stdin, os.Stdout, and
	// os.Stderr, respectively).
}

// displayName returns the flag name with optional short alias and type hint. When hasAnyShort is
// true, flags without a short alias are padded to align with those that have one. Examples: "-v,
// --verbose", "-o, --output string", "    --config string", "--debug".

// ParseToEnd is a drop-in replacement for flag.Parse. It improves upon the standard behavior by
// parsing flags even when they are interspersed with positional arguments. This overcomes Go's
// default limitation of stopping flag parsing upon encountering the first positional argument. For
// more details, see:
//
//   - https://github.com/golang/go/issues/4513
//   - https://github.com/golang/go/issues/63138
//
// This is a bit unfortunate, but most users nowadays consuming CLI tools expect this behavior.

// Exit codes:
//   - 0: successful completion
//   - 1: run function returned an error
//   - 124: shutdown timeout exceeded
//   - 130: forced shutdown (second signal or immediate termination)
//
// Example: HTTP server
//
//	server := &http.Server{
//	    Addr: ":8080",
//	    Handler: mux,
//	}
//
//	graceful.Run(
//	    graceful.ListenAndServe(server, 15*time.Second),       // HTTP draining period
//	    graceful.WithTerminationTimeout(30*time.Second),  // overall shutdown limit
//	)
