#!/bin/bash
set -e

/app/queryasnnames /app/friendlytag.db
export PATH=/app/:${PATH}

if [ "${1:0:1}" = '-' ]; then
    set -- telegraf "$@"
fi

exec "$@"
