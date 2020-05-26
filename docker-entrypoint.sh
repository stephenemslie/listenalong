#!/bin/bash

PATH="$HOME/.poetry/bin:$PATH"

if [ "$1" = 'runserver' ]; then
    poetry run ./manage.py migrate
    exec poetry run ./manage.py runserver 0:8000
fi

exec "$@"

