#!/usr/bin/sed -Enf

# nbconvert formats output blocks as idented sections of Markdown.
# We recognise these (with 4 spaces per indentation level),
# remove this leading whitespace, then attempt to infer the type of
# output as JSON, YAML, or simply a generic block.

# We're within an output block
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

  # If it's JSON
  /^\s*\{.*\}\s*$/ {
    i\```json
    b ENDBLOCK
  }

  # If it's YAML
  /.*:.*/ {
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
