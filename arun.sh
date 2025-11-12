#!/bin/bash

# Exit immediately if any command fails
set -e

# Colors for better readability
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Navigate to project root
cd "$(dirname "$0")"

echo -e "${GREEN}Starting Go backend...${NC}"
cd backend
go run . &
BACKEND_PID=$!

# Wait for processe
echo -e "${GREEN}Backend is running.${NC}"
echo "Press Ctrl+C to stop backend."

# Trap Ctrl+C and kill backend
trap "kill $BACKEND_PID" INT

wait
