run:
	go run ./cmd/web

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application 
.PHONY: build
build:
	@echo 'Building cmd/web...'
	GOOS=linux GOARCH=amd64 go build -o queue-app ./cmd/web

## docker/build: build docker image locally
.PHONY: docker/build
docker/build:
	@echo 'Building Docker image...'
	docker build -t queue-management-system .

## docker/run: run docker image locally
.PHONY: docker/run
docker/run:
	@echo 'Running Docker container...'
	docker run -p 4000:8080 -e PORT=8080 queue-management-system

## docker/test: build and test docker image locally
.PHONY: docker/test
docker/test: docker/build docker/run

# ==================================================================================== #
# DIGITALOCEAN DEPLOYMENT
# ==================================================================================== #

## deploy/create: create app on DigitalOcean (requires doctl)
.PHONY: deploy/create
deploy/create:
	@echo 'Creating app on DigitalOcean...'
	doctl apps create --spec .do/app.yaml

## deploy/update: update existing app on DigitalOcean
.PHONY: deploy/update
deploy/update:
	@echo 'Updating app on DigitalOcean...'
	doctl apps update $(APP_ID) --spec .do/app.yaml

## deploy/list: list all apps on DigitalOcean
.PHONY: deploy/list
deploy/list:
	@echo 'Listing apps on DigitalOcean...'
	doctl apps list

## deploy/logs: view app logs (requires APP_ID)
.PHONY: deploy/logs
deploy/logs:
	@echo 'Viewing app logs...'
	doctl apps logs $(APP_ID)

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

user = root
password = 933800Project
production_host_ip = 137.184.180.248

## production/connect: connect to the production server
.PHONY: connect
connect:
	ssh ${user}@${production_host_ip}
