PWD := `pwd`

default: build

build:
	docker run --rm -v $(PWD):/go/src/dp-docker-stats -w /go/src/dp-docker-stats golang:1.8-alpine /go/src/dp-docker-stats/hack/make.sh
	docker build -t dp-docker-stats .
clean:
	rm  -f ./dp-docker-stats
	docker rmi dp-docker-stats