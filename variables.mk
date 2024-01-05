## Binary
TARGET ?= simpleca

## Run additional parameters
RUN_PARAMETERS ?= web -dir . -port 8080

## Docker target image
DOCKER_IMAGE ?= $(TARGET)

## Docker target image tag
DOCKER_IMAGE_TAG ?= latest

## Docker run additional parameters
DOCKER_RUN_PARAMETERS ?= --publish 1443:443

## Tech-Lab Gitlab URL
GITLAB_URL ?=

## Tech-Lab Gitlab access token
GITLAB_TOKEN ?=

## Gitlab destination package
GITLAB_PACKAGE ?= $(DOCKER_IMAGE)

## Tech-Lab Nexus URL
NEXUS_URL ?=

## Nexus credentials user
NEXUS_AUTH_USER ?=

## Nexus credentials pass
NEXUS_AUTH_PASS ?=
