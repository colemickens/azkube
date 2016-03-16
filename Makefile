.NOTPARALLEL:

owner := colemickens
projectname := azkube
version := v0.0.2

imagename := $(owner)/$(projectname):$(version)
imagelatest := $(owner)/$(projectname):latest

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
	docker tag "$(imagename)" "$(imagelatest)"
	docker push "$(imagename)"
	docker push "$(imagelatest)"

clean:
	rm -f azkube

update-ca-certificates:
	#curl -o ca-certificates.crt http://www.cacert.org/certs/root.crt
	cp /etc/ssl/certs/ca-certificates.crt .
