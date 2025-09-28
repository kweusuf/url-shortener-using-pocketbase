#!/bin/sh

# Setup script for PocketBase admin initialization
# This script will be run on container startup to handle admin setup

PB_DATA_DIR="/app/pb_data"
PB_BINARY="./pocketbase-demo"

# Function to check if admin user exists
admin_exists() {
    # Check if there's at least one admin user in the _users collection
    if [ -f "$PB_DATA_DIR/data.db" ]; then
        # Use sqlite3 to check for admin users (role = 'admin')
        ADMIN_COUNT=$(sqlite3 "$PB_DATA_DIR/data.db" "SELECT COUNT(*) FROM _users WHERE role = 'admin';" 2>/dev/null || echo "0")
        [ "$ADMIN_COUNT" -gt 0 ] 2>/dev/null
    else
        return 1
    fi
}

# Check if this is the first run
if [ ! -d "$PB_DATA_DIR" ] || [ ! -f "$PB_DATA_DIR/data.db" ] || ! admin_exists; then
    echo "First run detected. Setting up PocketBase admin..."

    # Create data directory if it doesn't exist
    mkdir -p "$PB_DATA_DIR"

    echo "Starting PocketBase for admin setup..."
    echo "You will need to create an admin user through the web interface."
    echo "Open http://localhost:8090/_/ in your browser to set up the admin user."
    echo "After creating the admin user, the application will be ready to use."

    # Start PocketBase normally - it will prompt for admin creation via web interface
    exec $PB_BINARY serve --http=0.0.0.0:8090 --dir=$PB_DATA_DIR
else
    echo "PocketBase already initialized with admin user. Starting normally..."
    # Start PocketBase normally
    exec $PB_BINARY serve --http=0.0.0.0:8090 --dir=$PB_DATA_DIR
fi
