#!/bin/bash
set -e

echo "Updating apt and installing dependencies..."
sudo apt update
sudo apt install -y docker.io docker-compose-v2 curl wget tar

echo "Downloading and installing Go 1.25.0..."
wget -q https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
rm go1.25.0.linux-amd64.tar.gz

echo "Setting up Go in PATH..."
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
fi
export PATH=$PATH:/usr/local/go/bin

echo "Downloading and installing golang-migrate..."
curl -sL https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
rm LICENSE README.md || true

echo "Starting Docker service..."
sudo systemctl enable --now docker

echo "Starting PostgreSQL container..."
sudo docker compose up -d

echo "Waiting for PostgreSQL to be ready..."
sleep 5

echo "Running migrations..."
export DSN="postgres://users:tapirhorse@localhost/rentals?sslmode=disable"
migrate -path=./migrations -database=$DSN up

echo "Downloading Go modules..."
go mod download

echo "Setup complete! Please run 'source ~/.bashrc' or restart your terminal to use 'go'."
