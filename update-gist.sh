#!/bin/sh

# bash ./update-graph.sh

go run main.go | gist -f test-stats.md -u 102d08389c61e7f7293570bf38205d3b 
echo update
