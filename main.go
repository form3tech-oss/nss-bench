package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/juju/ratelimit"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
	"github.com/thanhpk/randstr"
	"go.etcd.io/etcd/pkg/report"
)

func main() {
	c := flag.String("c", "", "the channel in which to publish messages")
	cid := flag.String("cid", "", "the ID of the NATS Streaming cluster")
	d := flag.Duration("d", 1*time.Minute, "the duration of the benchmark")
	mpa := flag.Int("mpa", 512, "the maximum number of in-flight PUBACKs")
	mps := flag.Int("mps", 1, "the number of messages to publish per publisher per second")
	ms := flag.Int("ms", 4096, "the size of each messages in bytes")
	np := flag.Int("np", 1, "the number of publishers to use")
	s := flag.String("s", "nats://nats-streaming:4222", "the NATS server(s) to connect to")
	v := flag.String("v", log.InfoLevel.String(), "the level to use for logging")
	flag.Parse()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	if vv, err := log.ParseLevel(*v); err == nil {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
		log.SetLevel(vv)
	} else {
		log.Fatalf("%q is not a valid log level: %v", vv, err)
	}
	rand.Seed(time.Now().UnixNano())

	// Use a random channel name.
	ch := *c
	if ch == "" {
		ch = randstr.Hex(16)
	}

	var wpd sync.WaitGroup
	var wpr sync.WaitGroup
	wpd.Add(*np)
	wpr.Add(*np)
	r := report.NewReport("%4.4f")
	for id := 0; id < *np; id++ {
		go func(id int) {
			defer wpd.Done()
			runPublisher(stopCh, r, *s, *cid, id, ch, *mpa, *mps, *ms, *d, &wpr)
		}(id)
	}
	rc := r.Run()
	wpd.Wait()
	close(r.Results())
	fmt.Println(<-rc)
}

func runPublisher(stopCh chan os.Signal, r report.Report, s, cid string, id int, channel string, mpa, mps, ms int, duration time.Duration, wpr *sync.WaitGroup) int {
	i := fmt.Sprintf("nss-bench-%d", id)
	l := log.WithField("id", i)

	n, err := nats.Connect(s, nats.PingInterval(2*time.Second), nats.Name(cid), nats.MaxReconnects(-1), nats.ReconnectWait(0), nats.ReconnectBufSize(-1), nats.DisconnectErrHandler(func(c *nats.Conn, err error) {
		log.Errorf("Connection to NATS lost [%v]: %v", c.Servers(), err)
	}), nats.ReconnectHandler(func(c *nats.Conn) {
		log.Infof("Reconnecting to %v [%v]...", c.ConnectedUrl(), c.Servers())
	}), nats.DiscoveredServersHandler(func(c *nats.Conn) {
		log.Infof("Discovered NATS servers: %v", c.Servers())
	}))
	if err != nil {
		l.Fatalf("Failed to connect to NATS: %v", err)
	}

	connCh := make(chan struct{}, 1)
	connFn := func() stan.Conn {
		v, err := stan.Connect(cid, i, stan.Pings(2, 2), stan.MaxPubAcksInflight(mpa), stan.NatsConn(n), stan.SetConnectionLostHandler(func(_ stan.Conn, err error) {
			connCh <- struct{}{}
			l.Errorf("Connection to NATS Streaming lost: %v", err)
		}))
		if err != nil {
			connCh <- struct{}{}
			l.Errorf("Failed to connect to NATS Streaming: %v", err)
		}
		return v
	}
	c := connFn()
	defer func() {
		if c != nil {
			if err := c.Close(); err != nil {
				l.Errorf("Failed to close connection: %v", err)
			}
		}
		n.Close()
	}()

	wpr.Done()
	wpr.Wait()

	d := time.NewTimer(duration)
	defer d.Stop()
	b := ratelimit.NewBucketWithRate(float64(mps), int64(mps))
	t := time.Now()
	m := 0

loop:
	for {
		select {
		case <-d.C:
			break loop
		case <-stopCh:
			break loop
		case <-connCh:
			c = connFn()
		default:
			b.Wait(1)
			v := make([]byte, ms)
			if _, err = rand.Read(v); err != nil {
				r.Results() <- report.Result{Start: time.Now(), End: time.Now(), Err: err}
			} else {
				y := time.Now()
				err := c.Publish(channel, v)
				if err != nil {
					l.Errorf("Failed to publish message: %v", err)
				}
				m++
				r.Results() <- report.Result{Start: y, End: time.Now(), Err: err}
			}
		}
	}
	e := time.Since(t)
	l.Infof("Published %d messages in %f seconds (%f)", m, e.Seconds(), float64(m)/e.Seconds())
	return m
}
