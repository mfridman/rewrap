#!/bin/bash

# This is a very long shell script comment that
# should be rewrapped to fit within the specified
# column width properly.

# Short comment.

some_command --flag value

# Another comment block that has been wrapped at
# an awkward width and should be reflowed to fill
# lines properly.

if [ -f "file" ]; then
    # This is an indented comment inside an if
    # block that is too long and should be
    # rewrapped to fit the column width.
    echo "found"
fi
