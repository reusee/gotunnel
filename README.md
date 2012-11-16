```
install:
git clone git://github.com/reusee/gotunnel.git
cd gotunnel
git submodule update --init

cd server
edit config.go
go build
./server

cd ../client
edit config.go
go build
./client
```
