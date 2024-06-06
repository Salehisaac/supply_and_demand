#!/bin/bash


remove_cdrom_entry() {
  if grep -q 'cdrom' /etc/apt/sources.list; then
    echo "Removing cdrom entry from sources.list..."
    sudo sed -i '/cdrom/d' /etc/apt/sources.list
    sudo apt-get update
  fi
}

read -p "Enter your Telegram bot token: " BOT_TOKEN
read -p "Enter your Telegram user chat ID: " CHAT_ID

MYSQL_ROOT_PASSWORD="root_password"
MYSQL_DATABASE="my_database"
MYSQL_USER="root"
MYSQL_PASSWORD=$MYSQL_ROOT_PASSWORD

cat <<EOL > ../.env
BOT_TOKEN=$BOT_TOKEN
CHAT_ID=$CHAT_ID
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DB=$MYSQL_DATABASE
MYSQL_USER=$MYSQL_USER
MYSQL_PASSWORD=$MYSQL_PASSWORD
EOL


if ! [ -x "$(command -v docker)" ]; then
  echo "Docker is not installed. Installing Docker..."
  
  remove_cdrom_entry

  # Update package list and install prerequisites
  sudo apt-get update
  sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

  # Add Docker’s official GPG key
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

  # Set up the stable repository
  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
    $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

  # Update package list again and install Docker
  sudo apt-get update
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

  # Verify Docker installation
  if ! [ -x "$(command -v docker)" ]; then
    echo "Docker installation failed. Exiting."
    exit 1
  fi
else
  echo "Docker is already installed."
fi

echo "Running MySQL Docker container..."
docker run --name mysql -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e MYSQL_DATABASE=$MYSQL_DATABASE -p 3306:3306 -d mysql:latest

echo "Running Redis Docker container..."
docker run --name redis -p 6379:6379 -d redis:latest

echo "Waiting for MySQL to initialize..."
sleep 60  


cd "$(dirname "$0")"

if [ -f "../internal/database/migrations/0001_initial.sql" ]; then
  echo "Migrating 0001_initial.sql to MySQL..."
  docker cp ../internal/database/migrations/0001_initial.sql mysql:/tmp/0001_initial.sql
  docker exec mysql sh -c 'mysql -uroot -proot_password -D my_database < /tmp/0001_initial.sql'
  echo "Migration complete."
else
  echo "0001_initial.sql file not found. Skipping migration."
fi

CONFIRMATION_TOKEN=$(openssl rand -hex 16)

docker exec -it redis redis-cli SET confirmation_token $CONFIRMATION_TOKEN

echo -e "Here is your token: $CONFIRMATION_TOKEN\nSave it somewhere, you're going to need it!"

