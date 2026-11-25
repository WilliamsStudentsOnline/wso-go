#!/usr/bin/env bash
# Convert a Swagger 2.0 spec to OpenAPI 3, which oasdiff requires to compare schemas
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <swagger.yaml> <out.yaml>" >&2
  exit 2
fi

npx --yes swagger2openapi@7.0.8 --yaml "$1" -o "$2"
test -s "$2"
