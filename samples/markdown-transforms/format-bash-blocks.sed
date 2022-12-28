#!/usr/bin/sed -nf

# nbconvert treats all code blocks as being Python blocks.
#
# Whenever we see a Python block, we accumulate that entire block
# and apply a heuristic to determine if this should be a shell block.
# The heuristic is: if the block contains any lines starting with a ! character,
# then it's probably a shell block.

/^```python/ {
  h
  b END
}

/^```$/ {
  x
  # If any start-of-line plings, treat as bash block
  /\n!/ {
    s/^```python/```bash/
    s/\n!/\n/g
  }

  # Regardless of whether or not we've changed the block,
  # we print the block followed by the terminating delimiter.
  p
  z
  x
  p
  b END
}

# If we're accumulating a block, keep accumulating.
# Otherwise, just print the current line and continue.
x
/```python/ {
  x
  H
  b END
}

x
p

:END
