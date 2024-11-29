#!/bin/bash

dirnames=("foo" "bar" "buz" "qux" "quux" "corge" "grault" "garply" "waldo" "fred")

mkdir out

for dirname in "${dirnames[@]}"
do
    mkdir "out/$dirname"

    echo "$dirname" > "out/$dirname/$dirname.txt"
done
