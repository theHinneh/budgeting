#!/bin/bash

# Test script for authentication endpoints
# Make sure your server is running on localhost:8080

BASE_URL="http://localhost:8080"

echo "🧪 Testing Authentication Endpoints"
echo "=================================="

# Test 1: Create a new user
echo "1. Creating a new user..."
CREATE_USER_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/users" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "firstname": "Test",
    "lastname": "User",
    "password": "password123",
    "phone_number": "+1234567890"
  }')

echo "Create User Response: $CREATE_USER_RESPONSE"
echo ""

# Test 2: Login with the created user
echo "2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

echo "Login Response: $LOGIN_RESPONSE"
echo ""

# Extract access token and refresh token from login response
ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$ACCESS_TOKEN" ] && [ -n "$REFRESH_TOKEN" ]; then
    echo "✅ Access token extracted: ${ACCESS_TOKEN:0:20}..."
    echo "✅ Refresh token extracted: ${REFRESH_TOKEN:0:20}..."
    
    # Test 3: Get current user (protected endpoint)
    echo "3. Getting current user..."
    CURRENT_USER_RESPONSE=$(curl -s -X GET "$BASE_URL/v1/auth/me" \
      -H "Authorization: Bearer $ACCESS_TOKEN")
    
    echo "Current User Response: $CURRENT_USER_RESPONSE"
    echo ""
    
    # Test 4: Get user sessions
    echo "4. Getting user sessions..."
    SESSIONS_RESPONSE=$(curl -s -X GET "$BASE_URL/v1/auth/sessions" \
      -H "Authorization: Bearer $ACCESS_TOKEN")
    
    echo "Sessions Response: $SESSIONS_RESPONSE"
    echo ""
    
    # Test 5: Test token refresh
    echo "5. Testing token refresh..."
    REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/auth/refresh" \
      -H "Content-Type: application/json" \
      -d "{
        \"refresh_token\": \"$REFRESH_TOKEN\"
      }")
    
    echo "Refresh Response: $REFRESH_RESPONSE"
    echo ""
    
    # Test 6: Test logout
    echo "6. Testing logout..."
    LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/auth/logout" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"refresh_token\": \"$REFRESH_TOKEN\"
      }")
    
    echo "Logout Response: $LOGOUT_RESPONSE"
    echo ""
    
else
    echo "❌ Failed to extract tokens from login response"
fi

echo "✅ Authentication tests completed!"
