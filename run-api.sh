#!/usr/bin/env bash
set -euo pipefail
cd /home/asdf/api
set -a
source /home/asdf/api/env-api.env
set +a
exec go run ./cmd/server
