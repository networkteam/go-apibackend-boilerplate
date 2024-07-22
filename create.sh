#!env bash

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
rsync -av $DIR/ $target/ --exclude backend/api/graph/generated --exclude .idea --exclude .git --exclude .devbox --exclude .github --exclude docs/site --exclude backend/tmp/refresh-build

cd $target

echo "Replacing placeholders"

find . \( -type d -name .git -prune \) -o -type f -print0 | LC_ALL=C xargs -0 sed -i "s/mytld/$mytld/g; s/myvendor/$myvendor/g; s/myproject/$myproject/g"

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
See [./docs](docs) for more information.

## Development Quickstart

With Devbox:

    devbox services up

### Requirements

* [Devbox](https://www.jetify.com/devbox/docs/installing_devbox/)
EOF

echo "Creating Git repository"

git init -q
git add .
git add -f backend/tmp/.gitkeep
git commit -q -m "Initial commit"

echo "Done."
echo

echo "Run project via Devbox:"
echo
echo "    devbox services up"
