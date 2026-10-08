#!/bin/bash
set -e

echo "Starting Deployment..."

# Set Non-Interactive for Apt
export DEBIAN_FRONTEND=noninteractive

# 1. Install Dependencies
echo "Installing Dependencies..."
sudo apt-get update
sudo apt-get install -y -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" postgresql postgresql-contrib nginx certbot python3-certbot-nginx

# 2. Setup Database (Idempotent)
echo "Configuring Database..."
sudo -u postgres psql -c "CREATE DATABASE sello;" || true
sudo -u postgres psql -c "CREATE USER sello_user WITH PASSWORD 'sello123';" || true
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE sello TO sello_user;" || true
# Alter password just in case it already exists
sudo -u postgres psql -c "ALTER USER sello_user WITH PASSWORD 'sello123';"

# 3. Setup App
echo "Configuring App..."
chmod +x /home/ubuntu/sello-app
# Update .env with sed
sed -i 's/DB_USER=.*/DB_USER=sello_user/' /home/ubuntu/.env
sed -i 's/DB_PASSWORD=.*/DB_PASSWORD=sello123/' /home/ubuntu/.env
sed -i 's/DB_NAME=.*/DB_NAME=sello/' /home/ubuntu/.env
sed -i 's/DB_HOST=.*/DB_HOST=localhost/' /home/ubuntu/.env
sed -i 's/PORT=.*/PORT=8080/' /home/ubuntu/.env
# Append GIN_MODE if not exists
if ! grep -q "GIN_MODE" /home/ubuntu/.env; then
  echo "GIN_MODE=release" >> /home/ubuntu/.env
fi

# 4. Systemd
echo "Setting up Systemd..."
cat <<EOF | sudo tee /etc/systemd/system/sello.service
[Unit]
Description=Sello Backend
After=network.target postgresql.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu
ExecStart=/home/ubuntu/sello-app
Restart=always
EnvironmentFile=/home/ubuntu/.env

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable sello
sudo systemctl restart sello

# 5. Nginx
echo "Setting up Nginx..."
cat <<EOF | sudo tee /etc/nginx/sites-available/sello
server {
    server_name app.sello.id console.sello.id;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

sudo ln -sfn /etc/nginx/sites-available/sello /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl restart nginx

echo "Deployment Complete! Visit http://app.sello.id"
