#!/bin/sh
set -eu

if [ ! -f .env ]; then
  echo "Arquivo .env ausente. Copie .env.example e configure os valores." >&2
  exit 1
fi

set -a
. ./.env
set +a

exec go run ./cmd/birthday-reminder
