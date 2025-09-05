#!/bin/bash

# set username and password 
echo "Reading username and password from .env file..."
source .env

echo "Generating MongoDB keyfile..."
openssl rand -base64 764 > mongo-keyfile && chmod 400 mongo-keyfile
echo "MongoDB keyfile generated!"

# check if Variables are set
if [ -z "$MONGO_INITDB_ROOT_USERNAME" ] || [ -z "$MONGO_INITDB_ROOT_PASSWORD" ] || [ -z "$DATABASE_HOST" ]; then
  echo "Error: MONGO_INITDB_ROOT_USERNAME, MONGO_INITDB_ROOT_PASSWORD, and DATABASE_HOST must be set"
  exit 1
fi

echo "Creating MongoDB network..."
docker network create mongoCluster

echo "Starting MongoDB container..."
docker run -d  \
  -p 27017:27017 \
  --name mongo1 \
  --network mongoCluster \
  --network caddy \
  -e MONGO_INITDB_ROOT_USERNAME=$MONGO_INITDB_ROOT_USERNAME \
  -e MONGO_INITDB_ROOT_PASSWORD=$MONGO_INITDB_ROOT_PASSWORD \
  --label "caddy:$DATABASE_HOST" \
  --label "caddy.reverse_proxy:http://$DATABASE_HOST:27017" \
  -v ./mongo-keyfile:/data/keyfile \
  -v mongodata:/data/db \
  mongo:5 \
  mongod --bind_ip_all --replSet myReplicaSet --auth --keyFile /data/keyfile

echo "Waiting for MongoDB to start (10 seconds)..."
sleep 10

echo "Initializing replica set..."
docker exec -it mongo1 mongosh -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD --eval "rs.initiate({
 _id: \"myReplicaSet\",
 members: [
   { _id: 0, host: \"localhost:27017\" }
 ]
})"
# run this to change host resolving
docker exec -it mongo1 mongosh -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD --eval "cfg = rs.conf() 
cfg.members[0].host = \"$DATABASE_HOST:27017\"
rs.reconfig(cfg, { force: true })
"
echo "MongoDB setup complete!"


# run redis container as well
echo "Starting Redis container..."  
docker run  -d --name redis --net mongoCluster -p 6379:6379 -p 8001:8001 redis/redis-stack:latest
echo "Redis setup complete!"