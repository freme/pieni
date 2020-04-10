# pieni
small file server

# Using docker

## Build the docker image

    docker build -t pieni .

## Run the docker container

    docker run -ti --name pieni --rm --publish 3000:3001 pieni

## Open http://localhost:3000

# Using docker-compose

## Build and run the image
    docker-compose build && docker-compose up

## Open http://localhost:3030

- click
- enjoy
