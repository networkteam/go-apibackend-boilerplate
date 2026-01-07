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
rsync -av $DIR/ $target/ --exclude create.sh --exclude backend/api/graph/generated --exclude .idea --exclude .git --exclude .devbox --exclude .github --exclude docs/site --exclude backend/tmp/refresh-build --exclude .venv --exclude .DS_Store --exclude docs/TODO.md

cd $target

echo "Replacing placeholders"

find . \( -type d -name .git -prune \) -o -type f -print0 | LC_ALL=C xargs -0 sed -i "s/mytld/$mytld/g; s/myvendor/$myvendor/g; s/myproject/$myproject/g"

pushd backend

echo "Formatting Go code"

go fmt ./...

echo "Generating GraphQL API"
go generate ./api/graph/...

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
git commit -q -m "wip: created from boilerplate"

echo "Done."
echo

echo "Next steps:"
echo
echo "1. Adapt CLAUDE.md files with your project details:"
echo "   - CLAUDE.md (project overview, domain concepts, authentication)"
echo "   - backend/CLAUDE.md (module paths already replaced)"
echo
echo "2. Run project via Devbox:"
echo
echo "    devbox services up"
