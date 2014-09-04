
Setup
=====

The following instructions assume that you have a properly setup golang environment. Please review https://golang.org/doc/code.html for proper setup instructions.

Download apache tika's jar file 'tika-app-1.5.jar' http://tika.apache.org/download.html you will also need a working java runtime environment.

    go get github.com/FunnyMonkey/sally-tika
    # Adjust configuration in config.json located in $GOPATH/src/github.com/FunnyMonkey/sally-tika/config.json
    go run $GOPATH/src/github.com/FunnyMonkey/sally-tika/main.go --config=$GOPATH/src/github.com/FunnyMonkey/sally-tika/config.json

Assuming you do not receive any errors the golang webserver should be running on port 8080. http://127.0.0.1:8080
