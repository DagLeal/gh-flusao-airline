#!/bin/sh
set -eu
# O disco persistente pode ser montado pelo host com dono root.
mkdir -p /data
chown app:app /data
exec su-exec app /server "$@"
