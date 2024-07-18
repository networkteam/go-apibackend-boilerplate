#!/bin/sh
set -e

./wait-for ${POSTGRES_HOST}:5432 -- echo "Database reachable"

./myproject-ctl migrate up

if [ ${CREATE_FIXTURES} -eq 1 ]
then
    ./myproject-ctl fixtures import
fi

exec ./myproject-ctl server "$@"
