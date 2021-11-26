// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	log.SetPrefix("upbot: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		q := make(chan os.Signal, 1)
		signal.Notify(q, os.Interrupt, os.Kill)
		sig := <-q
		log.Printf("%v", sig)
		cancel()
	}()

	group := map[int][]*monitor{}
	group[conf.Default.Interval] = []*monitor{}

	for _, m := range conf.Monitors {
		i := conf.Default.Interval
		if m.Interval != 0 {
			i = m.Interval
		}
		m.Status = statusStart

		if _, ok := group[i]; ok {
			group[i] = append(group[i], m)
		} else {
			group[i] = []*monitor{m}
		}
	}

	go serve(ctx)

	wg := sync.WaitGroup{}
	wg.Add(len(group))
	for interval, ms := range group {
		_in := interval
		_ms := ms
		go func() {
			defer wg.Done()
			checker(ctx, _in, _ms)
		}()
	}
	wg.Wait()
}

func serve(ctx context.Context) {
	addr := os.Getenv("UPBOT_ADDR")
	if len(addr) == 0 {
		addr = conf.Addr
	}

	http.Handle(wrapper(app, "/upbot", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("I'm OK"))
	})))
	log.Printf("upbot is on: %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("failed to start upbot: %v", err)
	}
}

func checker(ctx context.Context, interval int, ms []*monitor) {
	t := time.NewTicker(time.Second * time.Duration(interval))
	for {
		select {
		case <-ctx.Done():
			log.Println("upbot is down, goodbye!")
			return
		case <-t.C:
			log.Println("checking...")
			wg := sync.WaitGroup{}
			for _, m := range ms {
				go func(m *monitor) {
					notifiers := []string{}
					notifiers = append(notifiers, conf.Default.Notifiers...)
					notifiers = append(notifiers, m.Notifiers...)

					err := m.CheckHealth(ctx, notifiers)
					if err != nil {
						log.Printf("%s check err: %v", m.Name, err)
					}
				}(m)
			}
			wg.Wait()
		}
	}
}
