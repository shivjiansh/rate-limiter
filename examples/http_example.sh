#!/usr/bin/env bash
set -euo pipefail

curl -H 'X-User-ID: demo' http://localhost:8080/v1/echo
