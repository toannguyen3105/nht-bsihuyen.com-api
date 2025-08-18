#!/bin/bash

# Exit ngay nếu có lỗi
set -e

# Đường dẫn tới thư mục dự án
PROJECT_DIR="/var/www/html/api.bsihuyen.diseasevault.cloud/backend"

echo "==> Di chuyển vào thư mục dự án"
cd "$PROJECT_DIR"

echo "==> Pull code"
git pull

echo "==> Build Go project"
GOOS=linux GOARCH=amd64 go build -o main main.go

echo "==> Restart systemd service"
sudo systemctl restart api.bsihuyen.diseasevault.cloud.service

echo "==> Done!"