y.go:
	goyacc koi.y

koi: y.go
	go build .

clean:
	rm -f koi y.go

test: koi
	make test/values

test/hello_world:
	./koi tests/hello_world/main.koi
	go run ./tests > tests/stdout.txt
	@diff tests/hello_world/stdout.txt tests/stdout.txt || (echo "test failed"; exit 1)

test/values:
	./koi tests/values/main.koi
	go run ./tests > tests/stdout.txt
	@diff tests/values/stdout.txt tests/stdout.txt || (echo "test failed"; exit 1)
