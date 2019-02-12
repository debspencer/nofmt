
all:
	go build

test:
	(cd parser && $(MAKE) test)
	go test

cover:
	go test -coverprofile cover.out

show:
	go test -coverprofile cover.out
	go tool cover -html=cover.out

