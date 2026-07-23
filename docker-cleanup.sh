#!/bin/bash

# Enhanced docker-cleanup.sh with error handling

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if containers are healthy
check_container_health() {
    print_status "Checking container health..."
    
    # Wait for containers to start
    sleep 5
    
    # Check if any containers are restarting or unhealthy
    failing_containers=$(docker compose ps --filter "status=restarting" --format "table {{.Name}}\t{{.Status}}" | tail -n +2)
    
    if [ ! -z "$failing_containers" ]; then
        print_error "Found containers that are restarting:"
        echo "$failing_containers"
        
        print_status "Showing logs for failing containers..."
        docker compose logs --tail=50
        
        return 1
    fi
    
    # Check for containers that exited with non-zero code
    exited_containers=$(docker compose ps -a --filter "status=exited" --format "table {{.Name}}\t{{.Status}}" | tail -n +2)
    
    if [ ! -z "$exited_containers" ]; then
        print_error "Found containers that exited:"
        echo "$exited_containers"
        
        print_status "Showing logs for exited containers..."
        docker compose logs --tail=50
        
        return 1
    fi
    
    print_status "All containers appear to be running normally"
    return 0
}

# Function to validate YAML files
validate_yaml_files() {
    # Skip validation if --skip-validation is passed
    if [ "$1" == "--skip-validation" ]; then
        print_warning "Skipping YAML validation"
        return 0
    fi
    
    print_status "Validating YAML configuration files..."
    
    # Check if docker-compose.yml is valid
    if ! docker compose config > /dev/null 2>&1; then
        print_error "docker-compose.yml has syntax errors!"
        docker compose config
        return 1
    fi
    
    # Check for common config files
    for config_file in "local.yml" "config.yml" "nakama.yml"; do
        if [ -f "$config_file" ]; then
            print_status "Validating $config_file..."
            
            # Show file info for debugging
            lines=$(wc -l < "$config_file")
            bytes=$(wc -c < "$config_file")
            print_status "File info: $lines lines, $bytes bytes"
            
            # Try to parse YAML with Python (if available and PyYAML is installed)
            if command -v python3 &> /dev/null && python3 -c "import yaml" &> /dev/null; then
                python_output=$(python3 -c "
import yaml
import sys
try:
    with open('$config_file', 'r') as f:
        content = f.read()
        yaml.safe_load(content)
        print('YAML is valid')
except Exception as e:
    print(f'YAML Error: {e}')
    sys.exit(1)
" 2>&1)
                
                if [ $? -ne 0 ]; then
                    print_error "$config_file has YAML syntax errors!"
                    print_error "Details: $python_output"
                    print_warning "File contents with line numbers:"
                    cat -n "$config_file"
                    print_warning "Hidden characters check:"
                    if [ "$(uname)" = "Darwin" ]; then
                        cat -v "$config_file"
                    else
                        cat -A "$config_file"
                    fi
                    return 1
                else
                    print_status "$config_file is valid"
                fi
            else
                print_warning "Python3 or PyYAML (python3 -c 'import yaml') not available, skipping Python YAML validation"
            fi
        fi
    done
    
    print_status "YAML validation complete"
    return 0
}

# Main script starts here
print_status "Starting Docker cleanup and restart process..."

# Check for skip validation flag
SKIP_VALIDATION=false
if [ "$1" == "--skip-validation" ] || [ "$2" == "--skip-validation" ]; then
    SKIP_VALIDATION=true
fi

# Validate YAML files before proceeding
if [ "$SKIP_VALIDATION" != "true" ]; then
    if ! validate_yaml_files; then
        print_error "YAML validation failed. Please fix configuration files before proceeding."
        print_status "You can skip validation with: ./docker-cleanup.sh --skip-validation"
        exit 1
    fi
else
    print_warning "Skipping YAML validation as requested"
fi

# Stop all containers and remove volumes
if [ "$1" == "-v" ] || [ "$2" == "-v" ]; then
    print_status "Stopping containers and removing volumes..."
    docker compose down -v
else
    print_status "Stopping containers..."
    docker compose down
fi

# Remove all images (with error handling)
print_status "Removing all Docker images..."
if docker images -q | xargs -r docker rmi 2>/dev/null; then
    print_status "Images removed successfully"
else
    print_warning "Some images could not be removed (may be in use)"
fi

# Remove any remaining containers (with error handling)
print_status "Removing any remaining containers..."
if docker ps -a -q | xargs -r docker rm 2>/dev/null; then
    print_status "Containers removed successfully"
else
    print_warning "No containers to remove or some could not be removed"
fi

# Pull latest images
print_status "Pulling latest images..."
if ! docker compose pull; then
    print_error "Failed to pull images"
    exit 1
fi

# Start containers
print_status "Starting containers..."
if ! docker compose up -d; then
    print_error "Failed to start containers"
    exit 1
fi

# Check container health (unless skipped)
if [ "$SKIP_VALIDATION" != "true" ]; then
    if ! check_container_health; then
        print_error "Container health check failed!"
        print_status "Here are the current container statuses:"
        docker compose ps
        
        print_status "Recent logs:"
        docker compose logs --tail=100
        
        print_error "Cleanup completed but containers are not healthy. Check the logs above."
        exit 1
    fi
else
    print_warning "Skipping health check as validation was skipped"
fi

print_status "Docker cleanup and restart complete! All containers are running normally."

# Show final status
docker compose ps