#!/bin/bash

# create a Docker network for the project
docker network create --driver bridge blog_engine_network

# Build containers
make docker-build
