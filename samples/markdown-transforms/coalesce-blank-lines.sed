#!/usr/bin/sed -zEnf

# Replace three or more new lines in a row with just two

s|(\n){3,}|\n\n|g
p
