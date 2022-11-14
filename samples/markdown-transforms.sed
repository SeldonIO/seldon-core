/^```python/ {
  h
  b END
}

/^```$/ {
  x
  s/```python/```bash/
  p
  x
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
