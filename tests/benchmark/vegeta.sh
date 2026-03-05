#!/usr/bin/env bash
set -euo pipefail

echo "GET http://localhost:8080/v1/echo" | vegeta attack -duration=30s -rate=100000 -workers=200 | vegeta report
