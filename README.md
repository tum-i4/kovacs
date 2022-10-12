# How to run this

## Important notice

- Do NOT rename the proofs of non-repudiation!

## Known issues

- Running memory or cpu profiler on listener when using fake chatter crashes the entire network

## Setup a bootnode

### Start a bootnode

This is done by navigating to the config/ directory and calling ```bootnode both```, which create and start a bootnode. After the bootnode is started, it will print it's enode to the console. Replace the local IP Address (127.0.0.1) after the '@' symbol with a network IP address, like "192.168.3.22".

### Allowing network access to the bootnode

The easiest way is configuring nginx to be a proxy. An example for this can be found in config/nginx_bootnode_config. The ports should not be changed! After copying the nginx config, run ```sudo systemctl restart nginx``` to restart nginx and apply the changes.

## Setup a Revolori instance

- Clone the [toolchain repo](https://github.com/tum-i4/inverse-transparency)
- Pull the `kovacs` branch
- Deploy Revolori in production mode as instructed in the README
    - Important: When running ```create-keys``` enter 'y' when asked if you want to create new keys! Otherwise Revolori may be missing the RSA keys which are required for the signing proccess.
- Setup up nginx to allow network access

## Setup this project

Run ```./setup``` and follow the prompts. This is only required when setting up the project for the first time or when you wish to alter the amount of nodes.



# Run this project

Simply run ```docker-compose up -d``` in this directory, after the setup was completed.

## Issue commands

Run ```docker exec -it inverse_transparency_node_1 sh``` to connect to node_1 and to run the requester.

## Run tests

Connect to container and enter ```export CGO_ENABLED=0``` and then run ```go test -bench=. -benchtime=10x```, where 10x can be replaced by any number.

Get cpu and memor usage with pprof ```go test -cpuprofile cpu.prof -memprofile mem.prof -bench .``` and view with ```go tool pprof cpu.prof``` or ```go tool pprof memory.prof```.

# Git hooks

git config core.hooksPath .githooks
