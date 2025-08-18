#!/bin/bash

# --- Configuration variables ---
DOMAIN="api.socmanga.diseasevault.cloud"
APP_NAME="$DOMAIN"
DOMAIN_NGINX_CONF="$DOMAIN.conf"
WEB_ROOT="/var/www/html/$DOMAIN/backend"
NGINX_CONF="/etc/nginx/sites-available/$DOMAIN_NGINX_CONF"
NGINX_LINK="/etc/nginx/sites-enabled/$DOMAIN_NGINX_CONF"
SERVICE_FILE="/etc/systemd/system/$APP_NAME.service"
ENV_FILE="$WEB_ROOT/app.env"

echo "⚠️ WARNING: This script will uninstall everything related to $DOMAIN"
read -p "Are you sure you want to continue? (yes/no): " confirm
if [[ "$confirm" != "yes" ]]; then
  echo "❌ Uninstall canceled."
  exit 1
fi

# --- Stop and disable systemd service ---
if systemctl is-active --quiet "$APP_NAME"; then
  echo "Stopping $APP_NAME service..."
  sudo systemctl stop "$APP_NAME"
fi

if systemctl is-enabled --quiet "$APP_NAME"; then
  echo "Disabling $APP_NAME service..."
  sudo systemctl disable "$APP_NAME"
fi

# --- Remove systemd service file ---
if [ -f "$SERVICE_FILE" ]; then
  echo "Removing service file: $SERVICE_FILE"
  sudo rm -f "$SERVICE_FILE"
  sudo systemctl daemon-reload
fi

# --- Remove nginx configuration ---
if [ -L "$NGINX_LINK" ]; then
  echo "Removing Nginx symlink: $NGINX_LINK"
  sudo rm -f "$NGINX_LINK"
fi

if [ -f "$NGINX_CONF" ]; then
  echo "Removing Nginx config: $NGINX_CONF"
  sudo rm -f "$NGINX_CONF"
fi

echo "Reloading Nginx..."
sudo nginx -t && sudo systemctl reload nginx

# --- Delete Let's Encrypt SSL cert ---
echo "Revoking and deleting SSL certificate using Certbot..."
sudo certbot delete --cert-name "$DOMAIN"

# --- Remove project files ---
if [ -d "$WEB_ROOT" ]; then
  echo "Removing project directory: $WEB_ROOT"
  sudo rm -rf "/var/www/html/$DOMAIN"
fi

# --- Done ---
echo "✅ Uninstallation completed for $DOMAIN"
