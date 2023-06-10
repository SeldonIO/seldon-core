#!/bin/bash

# Add Golang Copyright headers
for i in $(find ./apis ./components ./hodometer ./operator ./scheduler ./samples/examples ./tests/integration -name '*.go')  # or whatever other pattern...
do
  if ! grep -q Copyright $i
  then
    cat hack/boilerplate.go.txt $i >$i.new && mv $i.new $i
  fi
done

# Add Python Copyright headers
for i in $(find ./samples/examples -name '*.py')  # or whatever other pattern...
do
  if ! grep -q Copyright $i
  then
    cat hack/boilerplate.python.txt $i >$i.new && mv $i.new $i
  fi
done

# Add Kotlin Copyright headers
for i in $(find ./apis ./scheduler/data-flow -name '*.kt')  # or whatever other pattern...
do
  if ! grep -q Copyright $i
  then
    cat hack/boilerplate.go.txt $i >$i.new && mv $i.new $i
  fi
done
