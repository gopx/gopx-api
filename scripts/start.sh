#!/usr/bin/env bash

reflex -r '\.go$' -R '^vendor/' -s -- sh -c ./scripts/run.sh