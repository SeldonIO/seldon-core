#!/bin/bash
if [ $# -lt 1 ]; then
    echo "Usage: $0 <input_file> [output_file]"
    exit 1
fi

input_file="$1"
output_file="$2"

format_dependency() {
    dependency="$1"
    jar="$2"
    license="$3"
    url="$4"

    echo "Dependency: $name" >> "$output_file"
    echo "Jar:        $file" >> "$output_file"
    echo "License:    $license $url" >> "$output_file"
    echo "--------------------------------------------------------------------------------" >> "$output_file"
}

dependencies=$(jq -r '.dependencies | .[] | @base64' "$input_file")

for dep in $dependencies; do
    _jq() {
        echo "${dep}" | base64 --decode | jq -r ${1}
    }

    name=$(_jq '.name')
    file=$(_jq '.file')
    license_name=$(_jq '.licenses[0].name')
    license_url=$(_jq '.licenses[0].url')
    format_dependency "$name" "$file" "$license_name" "$license_url"
done
