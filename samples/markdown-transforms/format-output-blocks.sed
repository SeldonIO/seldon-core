#!/usr/bin/sed -Enf

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
