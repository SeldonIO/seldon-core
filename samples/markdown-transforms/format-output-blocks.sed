#!/usr/bin/sed -Enf

/^\s{4}.*/ {
  s/^\s{4}//
  H
  b END
}

x

# If we've accumulated an output block...
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

x
p

:END
