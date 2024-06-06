#!/bin/bash
set -uo pipefail

sudo sudo apt-get install -y bats

filesToCheck=$(find bash_tests -type f -name "*.test.bats.sh" )

echo $filesToCheck

bats -p ${filesToCheck}