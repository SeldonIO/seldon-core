#!/usr/bin/sed -Ef

# Remove trailing whitespace per line.

s|(\S*)\s+$|\1|
