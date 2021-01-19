#!/bin/bash
set -e

golangci-lint cache clean
golangci-lint run --max-same-issues 0