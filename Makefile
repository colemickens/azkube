.NOTPARALLEL:

owner := colemickens
projectname := azkube
version := v0.0.1

imagename := $(owner)/$(projectname):$(version)

all: build

glide:
	GO15VENDOREXPERIMENT=1 glide up
	GO15VENDOREXPERIMENT=1 glide rebuild

build:
	GO15VENDOREXPERIMENT=1 \
	CGO_ENABLED=0 \
	go build -a -tags netgo -installsuffix nocgo -ldflags '-w' .

docker: clean build
	docker build -t "$(imagename)" .

docker-push: docker
	docker push "$(imagename)"

clean:
	rm -f azkube
