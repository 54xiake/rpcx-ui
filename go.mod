module github.com/smallnest/rpcx-ui

go 1.14

require (
	github.com/54xiake/libkv v1.0.2
	github.com/docker/libkv v0.2.1
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.1
	github.com/smallnest/libkv-etcdv3-store v1.1.8
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
