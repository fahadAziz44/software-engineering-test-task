#!/bin/bash

# API Testing Script for User Management API
# Usage: ./test-api.sh [base_url] [--no-cleanup]
# Default base_url: http://localhost:8080
# --no-cleanup: Skip cleanup of test data (useful for debugging)

BASE_URL="${1:-http://localhost:8080}"
API_BASE="${BASE_URL}/api/v1"
NO_CLEANUP=false

# Check for --no-cleanup flag
for arg in "$@"; do
    if [ "$arg" == "--no-cleanup" ]; then
        NO_CLEANUP=true
    fi
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
PASS=0
FAIL=0

# Array to track created test users for cleanup
CREATED_USERS=()

print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_test() {
    echo -e "${YELLOW}TEST: $1${NC}"
}

print_pass() {
    echo -e "${GREEN}✓ PASS${NC}\n"
    ((PASS++))
}

print_fail() {
    echo -e "${RED}✗ FAIL: $1${NC}\n"
    ((FAIL++))
}

check_status() {
    local expected=$1
    local actual=$2
    local test_name=$3

    if [ "$actual" -eq "$expected" ]; then
        print_pass
    else
        print_fail "Expected status $expected, got $actual"
    fi
}

# Check if server is running
print_header "Checking Server Status"
if curl -s "${BASE_URL}" > /dev/null 2>&1; then
    echo -e "${GREEN}Server is running at ${BASE_URL}${NC}\n"
else
    echo -e "${RED}ERROR: Server is not running at ${BASE_URL}${NC}"
    echo "Please start the server first: make run"
    exit 1
fi

# =============================================================================
# GET /api/v1/users - Get All Users
# =============================================================================
print_header "Test: GET /api/v1/users - Get All Users"

print_test "Should return 200 and list of users"
RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE}/users")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 200 "$HTTP_CODE" "Get all users"

# =============================================================================
# GET /api/v1/users/id/:id - Get User by ID
# =============================================================================
print_header "Test: GET /api/v1/users/id/:id - Get User by ID"

# First, get a real UUID from the system by fetching jdoe
print_test "Getting valid UUID from existing user (jdoe)"
RESPONSE=$(curl -s "${API_BASE}/users/username/jdoe")
VALID_UUID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$VALID_UUID" ]; then
    echo -e "${RED}Warning: Could not get valid UUID from jdoe user${NC}"
    VALID_UUID="00000000-0000-0000-0000-000000000000"
fi

print_test "Should return 200 for valid user UUID ($VALID_UUID)"
RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE}/users/id/${VALID_UUID}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 200 "$HTTP_CODE" "Get user by valid UUID"

print_test "Should return 404 for non-existent UUID (valid format but doesn't exist)"
NON_EXISTENT_UUID="12345678-1234-1234-1234-123456789012"
RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE}/users/id/${NON_EXISTENT_UUID}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 404 "$HTTP_CODE" "Get user by non-existent UUID"

print_test "Should return 400 for invalid UUID format (not a UUID)"
RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE}/users/id/not-a-uuid")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Get user by invalid UUID format"

# =============================================================================
# GET /api/v1/users/username/:username - Get User by Username
# =============================================================================
print_header "Test: GET /api/v1/users/username/:username - Get User by Username"

print_test "Should return 200 for existing username (jdoe)"
RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE}/users/username/jdoe")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 200 "$HTTP_CODE" "Get user by existing username"

print_test "Should return 404 for non-existent username (nonexistent)"
RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE}/users/username/nonexistent")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 404 "$HTTP_CODE" "Get user by non-existent username"

# =============================================================================
# POST /api/v1/users - Create User (Success Cases)
# =============================================================================
print_header "Test: POST /api/v1/users - Create User (Success Cases)"

# Generate unique username for this test run
TIMESTAMP=$(date +%s)
TEST_USER="testuser${TIMESTAMP}"

print_test "Should create user with valid data"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${TEST_USER}\", \"email\": \"${TEST_USER}@example.com\", \"full_name\": \"Test User\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 201 "$HTTP_CODE" "Create user with valid data"
if [ "$HTTP_CODE" -eq 201 ]; then
    CREATED_USERS+=("${TEST_USER}")
fi

print_test "Should create user with apostrophe in name"
TEST_USER2="testuser${TIMESTAMP}b"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${TEST_USER2}\", \"email\": \"${TEST_USER2}@example.com\", \"full_name\": \"John O'Brien\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 201 "$HTTP_CODE" "Create user with apostrophe in name"
if [ "$HTTP_CODE" -eq 201 ]; then
    CREATED_USERS+=("${TEST_USER2}")
fi

print_test "Should create user with hyphen in name"
TEST_USER3="testuser${TIMESTAMP}c"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${TEST_USER3}\", \"email\": \"${TEST_USER3}@example.com\", \"full_name\": \"Mary-Jane Smith\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 201 "$HTTP_CODE" "Create user with hyphen in name"
if [ "$HTTP_CODE" -eq 201 ]; then
    CREATED_USERS+=("${TEST_USER3}")
fi

# =============================================================================
# POST /api/v1/users - Create User (Validation Errors)
# =============================================================================
print_header "Test: POST /api/v1/users - Validation Errors"

print_test "Should return 400 for missing username"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"test@example.com\", \"full_name\": \"Test User\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Missing username"

print_test "Should return 400 for missing email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"testuser\", \"full_name\": \"Test User\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Missing email"

print_test "Should return 400 for missing full_name"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"testuser\", \"email\": \"test@example.com\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Missing full_name"

print_test "Should return 400 for invalid email format"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"testuser\", \"email\": \"invalid-email\", \"full_name\": \"Test User\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Invalid email format"

print_test "Should return 400 for username too short (< 3 chars)"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"ab\", \"email\": \"test@example.com\", \"full_name\": \"Test User\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Username too short"

print_test "Should return 400 for username with special characters"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"test_user!\", \"email\": \"test@example.com\", \"full_name\": \"Test User\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Username with special characters"

print_test "Should return 400 for username with underscore"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"test_user\", \"email\": \"test@example.com\", \"full_name\": \"Test User\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Username with underscore"

# =============================================================================
# POST /api/v1/users - Create User (Conflict Errors)
# =============================================================================
print_header "Test: POST /api/v1/users - Conflict Errors"

print_test "Should return 409 for duplicate username"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${TEST_USER}\", \"email\": \"different@example.com\", \"full_name\": \"Different User\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 409 "$HTTP_CODE" "Duplicate username"

print_test "Should return 409 for duplicate email"
TEST_USER4="testuser${TIMESTAMP}d"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${TEST_USER4}\", \"email\": \"${TEST_USER}@example.com\", \"full_name\": \"Different User\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 409 "$HTTP_CODE" "Duplicate email"

# =============================================================================
# PATCH /api/v1/users/id/:id - Update User (Success Cases)
# =============================================================================
print_header "Test: PATCH /api/v1/users/id/:id - Update User (Success Cases)"

# Get UUID of first test user for update tests
RESPONSE=$(curl -s "${API_BASE}/users/username/${TEST_USER}")
UPDATE_USER_UUID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

print_test "Should update only full_name"
RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "${API_BASE}/users/id/${UPDATE_USER_UUID}" \
    -H "Content-Type: application/json" \
    -d '{"full_name": "Updated Test User"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 200 "$HTTP_CODE" "Update only full_name"

print_test "Should update only email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "${API_BASE}/users/id/${UPDATE_USER_UUID}" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"newemail${TIMESTAMP}@example.com\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 200 "$HTTP_CODE" "Update only email"

print_test "Should update multiple fields at once"
RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "${API_BASE}/users/id/${UPDATE_USER_UUID}" \
    -H "Content-Type: application/json" \
    -d '{"full_name": "Multi Update User", "email": "multiupdate'"${TIMESTAMP}"'@example.com"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 200 "$HTTP_CODE" "Update multiple fields"

# =============================================================================
# PATCH /api/v1/users/id/:id - Update User (Error Cases)
# =============================================================================
print_header "Test: PATCH /api/v1/users/id/:id - Update User (Error Cases)"

print_test "Should return 404 for non-existent UUID"
NON_EXISTENT_UUID="12345678-1234-1234-1234-123456789012"
RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "${API_BASE}/users/id/${NON_EXISTENT_UUID}" \
    -H "Content-Type: application/json" \
    -d '{"full_name": "Should Not Work"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 404 "$HTTP_CODE" "Update non-existent user"

print_test "Should return 400 for invalid UUID format"
RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "${API_BASE}/users/id/not-a-uuid" \
    -H "Content-Type: application/json" \
    -d '{"full_name": "Should Not Work"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Update with invalid UUID format"

print_test "Should return 400 for invalid email format"
RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "${API_BASE}/users/id/${UPDATE_USER_UUID}" \
    -H "Content-Type: application/json" \
    -d '{"email": "invalid-email"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Update with invalid email"

print_test "Should return 400 for username too short"
RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "${API_BASE}/users/id/${UPDATE_USER_UUID}" \
    -H "Content-Type: application/json" \
    -d '{"username": "ab"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Update with username too short"

print_test "Should return 409 for duplicate username"
RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "${API_BASE}/users/id/${UPDATE_USER_UUID}" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${TEST_USER2}\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 409 "$HTTP_CODE" "Update with duplicate username"

# =============================================================================
# DELETE /api/v1/users/id/:id - Delete User
# =============================================================================
print_header "Test: DELETE /api/v1/users/id/:id - Delete User"

# Create a user specifically for deletion test
TEST_DELETE_USER="testdelete${TIMESTAMP}"
RESPONSE=$(curl -s -X POST "${API_BASE}/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${TEST_DELETE_USER}\", \"email\": \"${TEST_DELETE_USER}@example.com\", \"full_name\": \"Delete Test User\"}")
DELETE_USER_UUID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

print_test "Should delete user with valid UUID (204 No Content)"
RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "${API_BASE}/users/id/${DELETE_USER_UUID}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 204 "$HTTP_CODE" "Delete user with valid UUID"

print_test "Should be idempotent - deleting again returns 204"
RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "${API_BASE}/users/id/${DELETE_USER_UUID}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 204 "$HTTP_CODE" "Delete already deleted user (idempotent)"

print_test "Should return 400 for invalid UUID format"
RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "${API_BASE}/users/id/not-a-uuid")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
check_status 400 "$HTTP_CODE" "Delete with invalid UUID format"

# =============================================================================
# Cleanup Test Data
# =============================================================================
if [ "$NO_CLEANUP" = false ] && [ ${#CREATED_USERS[@]} -gt 0 ]; then
    print_header "Cleaning Up Test Data"

    CLEANUP_SUCCESS=0
    CLEANUP_FAIL=0

    for username in "${CREATED_USERS[@]}"; do
        echo -e "${YELLOW}Deleting test user: ${username}${NC}"

        # Get user UUID by username
        USER_RESPONSE=$(curl -s "${API_BASE}/users/username/${username}")
        USER_UUID=$(echo "$USER_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

        if [ -n "$USER_UUID" ]; then
            # Delete using DELETE endpoint
            DELETE_RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "${API_BASE}/users/id/${USER_UUID}")
            DELETE_STATUS=$(echo "$DELETE_RESPONSE" | tail -n1)

            if [ "$DELETE_STATUS" -eq 204 ]; then
                echo -e "${GREEN}Successfully deleted ${username}${NC}"
                ((CLEANUP_SUCCESS++))
            else
                echo -e "${RED}Failed to delete ${username} (HTTP ${DELETE_STATUS})${NC}"
                ((CLEANUP_FAIL++))
            fi
        else
            echo -e "${RED}Could not find UUID for ${username}${NC}"
            ((CLEANUP_FAIL++))
        fi
    done

    echo -e "\n${GREEN}Cleanup complete: ${CLEANUP_SUCCESS} users deleted${NC}"
    if [ $CLEANUP_FAIL -gt 0 ]; then
        echo -e "${RED}Cleanup failed for ${CLEANUP_FAIL} users${NC}"
        echo -e "\n${BLUE}To manually clean up remaining users via database:${NC}"
        echo -e "docker exec database psql -U postgres -c \"DELETE FROM users WHERE username LIKE 'testuser%';\""
    fi
else
    if [ "$NO_CLEANUP" = true ]; then
        print_header "Cleanup Skipped"
        echo -e "${YELLOW}Test data was not cleaned up (--no-cleanup flag used)${NC}"
        if [ ${#CREATED_USERS[@]} -gt 0 ]; then
            echo -e "\nTest users created:"
            for username in "${CREATED_USERS[@]}"; do
                echo "  - $username"
            done
        fi
    fi
fi

# =============================================================================
# Summary
# =============================================================================
print_header "Test Summary"

TOTAL=$((PASS + FAIL))
echo -e "Total Tests: ${TOTAL}"
echo -e "${GREEN}Passed: ${PASS}${NC}"
echo -e "${RED}Failed: ${FAIL}${NC}\n"

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed! ✗${NC}"
    exit 1
fi
