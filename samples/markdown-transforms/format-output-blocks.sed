#!/usr/bin/sed -Enf

# nbconvert formats output blocks as idented sections of Markdown.
# We recognise these (with 4 spaces per indentation level),
# remove this leading whitespace, then attempt to infer the type of
# output as JSON, YAML, or simply a generic block.

################################################################################

# Ignore indentation within code blocks.

# If we find a starting delimiter, record that we're in a code block.
/^```.+/ {
    h
    p
    b END
}

# If we find an ending delimiter, stop acting like we're in a code block.
/^```$/ {
    p
    z
    h
    b END
}

# Check if we're in a code block.
# If so, print this line and skip to the next.
x
/^```/ {
    x
    p
    b END
}

# Otherwise, keep investigating the current line.
x

################################################################################

# If we're within an output block...
/^\s{4}.*/ {
  s/^\s{4}//
  H
  b END
}

# We're outside an output block
x

# If we've encountered an output block...
/^.+$/ {
  # Remove any lingering whitespace
  s/^\s+//
  s/\s+$//

  # Remove ANSI colour codes
  s/\x1B\[[0-9;]*m//g

  # If it's JSON
  /^\s*\{.*\}\s*$/ {
    i\```json
    b ENDBLOCK
  }

  # If it's YAML
  /^\w+:[^:\n]*\n/ {
    i\```yaml
    b ENDBLOCK
  }

  # Otherwise
  i\```

  :ENDBLOCK
  a\```\n
  p
  z
}

# Print whatever line we were on
x
p

:END
