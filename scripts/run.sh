#!/usr/bin/env bash

# Export the necessary environment variables
export $(cat scripts/.env | xargs)

# Install the API server executable
go install ./cmd/gopx-api

# Run the server
gopx-api