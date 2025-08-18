#!/bin/bash

set -e  # Dừng script ngay khi có lỗi

APP_PATH="/home/tony/Backend/simplebank/app"
APP_ENV_PATH="/home/tony/Backend/simplebank/app.env"
SERVER_USER="toannh"
SERVER_IP="64.176.82.62"
SERVER_PATH="/var/www/html/simplebank.codingwithtony.io.vn/backend/"
SERVICE_NAME="simplebank-api"

# Xóa file app nếu tồn tại
if [ -f "$APP_PATH" ]; then
  echo "Đang xóa file app cũ..."
  rm -f "$APP_PATH"
  echo "File app cũ đã bị xóa."
else
  echo "Không tìm thấy file app cũ, tiếp tục build."
fi

# Build dự án bằng Go
echo "Đang build dự án với GO..."
GOOS=linux GOARCH=amd64 go build -o "$APP_PATH"

echo "Build thành công! Đang copy file app lên server..."

# Copy file app, app.env lên server
rsync -avz --progress "$APP_PATH" "$APP_ENV_PATH" "$SERVER_USER@$SERVER_IP:$SERVER_PATH"

echo "Copy thành công! Đang kết nối tới server để restart service..."

# Restart service trên server