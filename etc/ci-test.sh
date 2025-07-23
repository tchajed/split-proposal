#!/bin/bash

set -eu

# assumes main.pdf is already built (using Makefile locally, or with github-action-for-latex)

out=$(go run . -file sample/main.pdf -outDir sample)
echo "$out" | grep "project-description: 2-5" 1>/dev/null
echo "$out" | grep "data-mgmt-plan: 7-7" 1>/dev/null

if [ ! -e sample/submit-project-description.pdf ]; then
  echo "project description not generated" 1>&2
  exit 1
fi
if [ ! -e sample/submit-mentoring-plan.pdf ]; then
  echo "mentoring plan not generated" 1>&2
  exit 1
fi

echo "passed"
