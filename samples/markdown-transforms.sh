#!/usr/bin/env bash

# Single pass with multiple transformations for efficiency
cat $1 \
  | ./markdown-transforms/format-bash-blocks.sed \
  | ./markdown-transforms/format-output-blocks.sed \
  | ./markdown-transforms/coalesce-blank-lines.sed \
  | ./markdown-transforms/remove-trailing-whitespace.sed \
  > $1.tmp

mv $1.tmp $1
