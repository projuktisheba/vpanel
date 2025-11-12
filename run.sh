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

echo -e "${GREEN}Starting React frontend...${NC}"
cd ../frontend
npm run dev &
FRONTEND_PID=$!

# Wait for both processes
echo -e "${GREEN}Both backend and frontend are running.${NC}"
echo "Press Ctrl+C to stop both."

# Trap Ctrl+C and kill both
trap "kill $BACKEND_PID $FRONTEND_PID" INT

wait
