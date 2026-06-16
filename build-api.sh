#!/usr/bin/env bash
set -euo pipefail
cd /home/asdf/api
mkdir -p /home/asdf/api/bin
go build -o /home/asdf/api/bin/prsnlspc-api ./cmd/server
