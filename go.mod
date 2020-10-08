module github.com/form3tech/nss-bench

go 1.13

replace go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200425165423-262c93980547

require (
	github.com/juju/ratelimit v1.0.1
	github.com/nats-io/nats-server/v2 v2.1.6 // indirect
	github.com/nats-io/nats-streaming-server v0.17.0 // indirect
	github.com/nats-io/nats.go v1.9.2
	github.com/nats-io/stan.go v0.6.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/sirupsen/logrus v1.5.0
	github.com/thanhpk/randstr v1.0.4
	go.etcd.io/etcd v3.3.20+incompatible
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)
