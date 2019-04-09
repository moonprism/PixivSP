build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o PixivSP .
	cp PixivSP docker/app
	rm PixivSP
	cp -r conf/*.ini docker/app/conf
	docker-compose build
	rm docker/app/conf/*.ini

init:
	docker-compose up mysql