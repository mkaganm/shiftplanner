#!/bin/bash
# AWS'de dosya yapısını kontrol etmek için script

echo "=== Root directory ==="
ls -la ~/shiftplanner/

echo ""
echo "=== Backend directory ==="
ls -la ~/shiftplanner/backend/ 2>/dev/null || echo "backend/ NOT FOUND"

echo ""
echo "=== Backend/cmd directory ==="
ls -la ~/shiftplanner/backend/cmd/ 2>/dev/null || echo "backend/cmd/ NOT FOUND"

echo ""
echo "=== Backend/cmd/server directory ==="
ls -la ~/shiftplanner/backend/cmd/server/ 2>/dev/null || echo "backend/cmd/server/ NOT FOUND"

echo ""
echo "=== Finding all cmd directories ==="
find ~/shiftplanner -type d -name "cmd" 2>/dev/null

echo ""
echo "=== Finding main.go ==="
find ~/shiftplanner -name "main.go" 2>/dev/null

