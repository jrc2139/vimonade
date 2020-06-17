#!/usr/bin/env bash

echo Enter the user@host to scp
read varname
scp "$1" "$varname":

ssh $varname zsh -c "'mv $1 ~/go/bin/'"

rm -f $1
