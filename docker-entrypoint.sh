#!/bin/bash

PATH="$HOME/.poetry/bin:$PATH"

if [ "$1" = 'runserver' ]; then
    poetry run ./manage.py migrate
    exec poetry run ./manage.py runserver 0:8000
fi

if [ "$1" = 'scheduler' ]; then
    export PYTHONUNBUFFERED=1
    exec poetry run ./manage.py start_scheduler
fi

exec "$@"

