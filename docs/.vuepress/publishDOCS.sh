#!/usr/bin/env sh

# abort on errors
set -e

./docs/.vuepress/moveCHANGELOG.sh

# build
npm run docs:build

# navigate into the build output directory
cd docs/.vuepress/dist


git init
git add -A
git commit -m 'deploy'

git push -f git@github.com:camomillacms/camomilla-core.git gh-pages

cd -