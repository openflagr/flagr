default: vet test

vet:
	go vet .

test:
	go test .

test-race:
	go test . -race

bench:
	go test . -run=NONE -bench=.
