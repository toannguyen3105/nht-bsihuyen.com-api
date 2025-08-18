#!/bin/bash

# --- Configuration variables ---
DOMAIN="api.socmanga.diseasevault.cloud"
# DOMAIN="api.socmanga.com"
DOMAIN_NGINX_CONF="$DOMAIN.conf"
EMAIL="nguyenhuytoan1994@gmail.com"
WEB_ROOT="/var/www/html/$DOMAIN/backend"
NGINX_CONF="/etc/nginx/sites-available/$DOMAIN_NGINX_CONF"
NGINX_LINK="/etc/nginx/sites-enabled/$DOMAIN_NGINX_CONF"
GITHUB_REPO="git@github.com:toannguyen3105/nht-bsihuyen.com-api.git"
APP_NAME="$DOMAIN"
PORT=9001
ENV_FILE="app.env"

# --- Clone the GOLANG project from GitHub ---
echo "Cloning the GOLANG project from GitHub into $WEB_ROOT..."
if [ -d "$WEB_ROOT" ]; then
  echo "Directory $WEB_ROOT already exists. Skipping git clone."
else
  echo "Cloning the GOLANG project from GitHub into $WEB_ROOT..."
  git clone "$GITHUB_REPO" "$WEB_ROOT"
fi

# --- Set ownership ---
echo "Setting ownership of /var/www/html/$DOMAIN to user $USER..."
sudo chown -R "$USER:$USER" "/var/www/html/$DOMAIN"
sudo chmod -R 755 /var/www/html/

echo "Navigating to the project directory and installing dependencies..."
cd "$WEB_ROOT" || {
  echo "Failed to change directory! Exiting script."
  exit 1
}

# Check if the .contentlayer directory exists
if [ -f main ]; then
  echo "Removing main file..."
  rm -rf main
else
  echo "main file does not exist."
fi

# --- Step to prompt for .env input ---
echo "Please enter your environment variables in the app.env file."

# Check if .env file exists or prompt the user to create one
if [ ! -f "$ENV_FILE" ]; then
    echo "Creating app.env file..."
    touch "$ENV_FILE"
fi

echo "You can now edit the app.env file manually. When you're done, type 'yes' to continue."
while true; do
    read -p "Have you finished editing the app.env file? (yes/no): " yn
    case $yn in
        [Yy]* ) echo "Proceeding with the build..."; break;;
        [Nn]* ) echo "Please finish editing the app.env file."; continue;;
        * ) echo "Please answer yes or no."; continue;;
    esac
done

# --- Build the GOLANG project ---
# Build binary
if ! GOOS=linux GOARCH=amd64 go build -o main main.go; then
  echo "❌ Build failed! Exiting script."
  exit 1
fi

echo "✅ Build successful."

# --- Configure Nginx ---
echo "Configuring Nginx..."

if [ ! -f "$NGINX_CONF" ]; then
  echo "Creating Nginx configuration at $NGINX_CONF..."
  sudo tee "$NGINX_CONF" > /dev/null <<EOF
server {
    listen 80;
    server_name $DOMAIN www.$DOMAIN;

    location / {
        proxy_pass http://localhost:$PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
EOF
else
  echo "Nginx config already exists at $NGINX_CONF. Skipping creation."
fi

if [ ! -L "$NGINX_LINK" ]; then
  echo "Linking $NGINX_CONF to sites-enabled..."
  sudo ln -s "$NGINX_CONF" "$NGINX_LINK"
fi

echo "Testing and reloading Nginx..."
sudo nginx -t && sudo systemctl reload nginx

# --- Create systemd service ---
SERVICE_FILE="/etc/systemd/system/$APP_NAME.service"

if [ -f "$SERVICE_FILE" ]; then
  echo "Systemd service already exists or forced creation. Restarting the service..."
  sudo systemctl restart "$APP_NAME"
else
  echo "Creating systemd service file at $SERVICE_FILE..."

  sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=Golang App - $APP_NAME
After=network.target

[Service]
User=$USER
WorkingDirectory=$WEB_ROOT
ExecStart=$WEB_ROOT/main
Restart=always

[Install]
WantedBy=multi-user.target
EOF

  echo "Reloading systemd to recognize the new service..."
  sudo systemctl daemon-reexec
  sudo systemctl daemon-reload

  echo "Enabling $APP_NAME service..."
  if ! sudo systemctl enable "$APP_NAME"; then
    echo "❌ Failed to enable $APP_NAME"
    sudo systemctl status "$APP_NAME"
    sudo journalctl -u "$APP_NAME" --no-pager -n 20
    exit 1
  fi

  echo "Restarting $APP_NAME service..."
  sudo systemctl restart "$APP_NAME"
  sleep 1

  if ! systemctl is-active --quiet "$APP_NAME"; then
    echo "❌ $APP_NAME failed to start"
    sudo systemctl status "$APP_NAME"
    sudo journalctl -u "$APP_NAME" --no-pager -n 20
    exit 1
  fi

  echo "✅ $APP_NAME is running successfully!"
fi

# --- Install Certbot if missing ---
if ! command -v certbot > /dev/null; then
    echo "Certbot not found. Installing certbot..."
    sudo apt update
    sudo apt install -y certbot python3-certbot-nginx
fi

# --- Obtain SSL certificate ---
echo "Requesting Let's Encrypt SSL certificate for $DOMAIN..."
sudo certbot --nginx -d "$DOMAIN" -d "www.$DOMAIN" --non-interactive --agree-tos -m "$EMAIL" --redirect