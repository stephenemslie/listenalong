#!/bin/bash

if [ "$1" = 'runserver' ]; then
    exec $HOME/.poetry/bin/poetry run ./manage.py runserver 0:8000
fi

$HOME/.poetry/bin/poetry shell

exec "$@"

