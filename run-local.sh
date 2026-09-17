#!/bin/sh
set -eu

if [ ! -f .env ]; then
  echo "Arquivo .env ausente. Copie .env.example e configure os valores." >&2
  exit 1
fi

while IFS='=' read -r key value || [ -n "$key" ]; do
  case "$key" in
    ''|'#'*) continue ;;
    [0-9]*|*[!A-Za-z0-9_]*)
      echo "Chave invalida no arquivo .env: $key" >&2
      exit 1
      ;;
  esac
  export "$key=$value"
done < .env

exec go run ./cmd/birthday-reminder
