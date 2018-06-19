#!/usr/bin/env bash

# Export the necessary environment variables
export $(cat scripts/.env | xargs)

# Install the Registry server executable
go install ./cmd/gopx-registry

# Run the server
gopx-registry