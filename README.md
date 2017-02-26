# Go File [![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/micro/go-file?status.svg)](https://godoc.org/github.com/micro/go-file)

Go File is a file server library leveraging go-micro. It enables you to serve and consume files via RPC.

This is a stripped down version of [gotransfer](https://github.com/yanolab/gotransfer).

## Usage


### Server

```go
import "github.com/micro/go-file"

service := micro.NewService(
	micro.Name("go.micro.srv.file"),
)

proto.RegisterFileHandler(service.Server(), file.NewHandler("/tmp"))

service.Init()
service.Run()
```

### Client

```go
import "github.com/micro/go-file"

// use new service or default client
service := micro.NewService()
service.Init()

client := file.NewClient("go.micro.srv.file", service.Client())
client.Download("remote.file", "local.file")
```
