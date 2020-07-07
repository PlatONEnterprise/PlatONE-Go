# test coverage rate

command demonstration:

running the following commands in ./ctool directory

test all the testing files:

`go test -v ./... -coverprofile=./test/coverage/test.out`

generate coverage report:

`go tool cover -html=./test/coverage/test.out -o ./test/coverage/test.html`
