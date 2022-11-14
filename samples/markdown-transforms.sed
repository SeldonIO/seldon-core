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
  p
  g
  p
  b END
}

x

/```python/ {
  x
  H
  b END
}

x

p

:END
