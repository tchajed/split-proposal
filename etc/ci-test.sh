#!/bin/bash

set -eu

# assumes main.pdf is already built (using Makefile locally, or with github-action-for-latex)

out=$(go run . -file sample/main.pdf -outDir sample)
echo "$out" | grep "project-description.pdf: p2.*4 pages" 1>/dev/null
echo "$out" | grep "data-mgmt-plan.pdf: p7" 1>/dev/null
echo "$out" | grep "Tej-Chajed-synergistic-activities.pdf: p9" 1>/dev/null

if [ ! -e sample/submit-project-description.pdf ]; then
  echo "project description not generated" 1>&2
  exit 1
fi
if [ ! -e sample/submit-mentoring-plan.pdf ]; then
  echo "mentoring plan not generated" 1>&2
  exit 1
fi
if [ ! -e sample/submit-Tej-Chajed-synergistic-activities.pdf ]; then
  echo "synergistic activities not generated" 1>&2
  exit 1
fi

echo "passed"
