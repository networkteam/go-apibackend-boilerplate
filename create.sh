#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [ "$#" -ne 4 ]; then
    echo "Usage: ./create.sh [target] [mytld] [myvendor] [myproject]"
    exit 1
fi

target=$1
mytld=$2
myvendor=$3
myproject=$4

echo "Copy $DIR to $target"

mkdir -p $target
rsync -av $DIR/ $target/ --exclude backend/api/graph/generated --exclude .idea --exclude .git

cd $target

echo "Replacing placeholders"

find . \( -type d -name .git -prune \) -o -type f -print0 | LC_ALL=C xargs -0 sed -i '' \
  "s/mytld/$mytld/g; s/myvendor/$myvendor/g; s/myproject/$myproject/g"

pushd backend

echo "Formatting Go code"

go fmt ./...

echo "Generating GraphQL API"
go run github.com/99designs/gqlgen

popd

echo "Creating README.md"
echo <<EOF > README.md
# $myproject

This project was kickstarted by go-apibackend-boilerplate.

EOF

echo "Creating Git repository"

git init -q
git add .
git add -f backend/tmp/.gitkeep
git commit -q -m "Initial commit"

echo "Done."
echo

echo "Run the following commands to create test and dev databases:"
echo
echo "    createdb $myproject-dev"
echo "    createdb $myproject-test"
echo
echo "Run migrations:"
echo
echo "    go run ./cli/ctl migrate up"
echo
echo "Run a server for development:"
echo
echo "    go run github.com/networkteam/refresh"
