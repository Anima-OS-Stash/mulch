#!/bin/bash

. ~/env

# move this definition to cloudinit generated env
source="/mnt/backup"

cd "$HTML_DIR" || exit $?
find "$source" || exit $?
