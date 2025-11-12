#!/bin/bash

# Exit immediately if any command fails
set -e

# Colors for better readability
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Navigate to project root
cd "$(dirname "$0")"

echo -e "${GREEN}Starting React frontend...${NC}"
cd frontend
npm run dev &
FRONTEND_PID=$!

# Wait for both processes
echo -e "${GREEN}Frontend is running.${NC}"
echo "Press Ctrl+C to stop frontend."

# Trap Ctrl+C and kill frontend
trap "kill $FRONTEND_PID" INT

wait
