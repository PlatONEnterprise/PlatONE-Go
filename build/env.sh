#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"

echo "$root" "$workspace"

bcosdir="$workspace/src/github.com/BCOSnetwork"
if [ ! -L "$bcosdir/BCOS-Go" ]; then
    mkdir -p "$bcosdir"
    cd "$bcosdir"
    ln -s ../../../../../. BCOS-Go
    cd "$root"
fi

echo "ln -s success."

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$bcosdir/BCOS-Go"
PWD="$bcosdir/BCOS-Go"

# Launch the arguments with the configured environment.
exec "$@"
