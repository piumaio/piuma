#!/bin/bash
destination="./docs"
destination=$(cd -- "$destination" && pwd)

# Move CHANGELOG 
changelog_dir="$destination/Changelog/"
mkdir -p $changelog_dir
cp ./CHANGELOG.md $changelog_dir/README.md

# Move CONTRIBUTING
contributing_dir="$destination/Contribute/"
mkdir -p $contributing_dir
cp ./CONTRIBUTING.md $contributing_dir/README.md

# Move LICENSE
license_dir="$destination/License/"
mkdir -p $license_dir
cp ./LICENSE $license_dir/README.md