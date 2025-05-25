#!/bin/bash

# down and remove all containers
make docker-down

# remove the Docker network
docker network rm blog_engine_network
