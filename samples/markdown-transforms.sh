

# convert python blocks to bash for all blocks
sed -zi 's/```python\n!/```bash\n/g' $1
# remove pling from seldon commands
sed -zri 's/!seldon/seldon/g' $1
# remove pling from seldon commands
sed -zri 's/!cat/cat/g' $1
# remove pling from kubectl commands
sed -zri 's/!kubectl/kubectl/g' $1
# remove pling from echo commands
sed -zri 's/!echo/echo/g' $1
# After a cat yaml command add a yaml block
sed -zri 's/cat([^\n]*)\n```\n/cat\1\n```\n````\{collapse\} Expand to see output\n```yaml/g' $1
# After a MESH_IP comand add bash block
sed -zri 's/MESH_IP([^\n]*)\n```\n/MESH_IP\1\n```\n````\{collapse\} Expand to see output\n```bash/g' $1
# Close blocks after indented areas
sed -zri 's/([ ]{3}[^\n]*\n)\n/\1```\n````/g'  $1
# After a bash seldon block add a json block
sed -zri 's/(```bash\nseldon[^`]*```)/\1\n````\{collapse\} Expand to see output\n```json/g' $1
# After a kubectl block add a json block
sed -zri 's/(```bash\nkubectl[^`]*```)/\1\n````\{collapse\} Expand to see output\n```json/g' $1
