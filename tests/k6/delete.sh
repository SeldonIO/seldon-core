#!/bin/bash
# Basic while loop
counter=1
while [ $counter -le 20000 ]
do
echo $counter
seldon pipeline unload tfsimplea$counter-pipeline --scheduler-host 34.90.84.16:9004
((counter++))
done
echo All done
