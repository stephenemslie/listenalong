#!/bin/sh

if [ "$1" = 'gowatch' ]; then
    exec CompileDaemon --build="go build -o /usr/src/app/bin/listenalong" --include="*.html" --command=./bin/listenalong
fi

exec "$@"
