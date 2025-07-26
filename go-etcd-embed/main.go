package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func main() {
	cfg := embed.NewConfig()
	cfg.Dir = "default.etcd"
	cfg.LogOutputs = []string{"./etcd.log"}
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer e.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	select {
	case <-e.Server.ReadyNotify():
		log.Println("Server is ready")
		if err := run(ctx); err != nil {
			log.Fatal(err)
		}
		e.Server.Stop()
		e.Close()
	case <-time.After(60 * time.Second):
		e.Server.Stop()
		log.Println("Server took too long to start")
	}

	log.Fatal(<-e.Err())
}

func run(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		if err := watcher(ctx); err != nil {
			log.Println(err)
		}
	}()
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if err := worker(ctx, "worker1"); err != nil {
			log.Println(err)
		}
	}()
	time.Sleep(time.Second)
	go func() {
		defer wg.Done()
		if err := worker(ctx, "worker2"); err != nil {
			log.Println(err)
		}
	}()

	wg.Wait()

	return nil
}

func worker(ctx context.Context, name string) error {
	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		return err
	}
	defer cli.Close()

	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			txn := cli.Txn(ctx)
			cmp := clientv3.Compare(clientv3.CreateRevision("leader"), "=", 0)
			grant, err := cli.Grant(ctx, 10)
			if err != nil {
				return err
			}
			putOp := clientv3.OpPut("leader", name, clientv3.WithLease(grant.ID))

			resp, err := txn.If(cmp).Then(putOp).Else(clientv3.OpGet("leader")).Commit()
			if err != nil {
				return err
			}

			var isLeader bool
			if !resp.Succeeded {

				for _, res := range resp.Responses {
					if rangeResp := res.GetResponseRange(); rangeResp != nil && len(rangeResp.Kvs) > 0 {
						key := string(rangeResp.Kvs[0].Key)
						value := string(rangeResp.Kvs[0].Value)
						if key != "leader" {
							continue
						}
						if value == name {
							isLeader = true
							break
						}
					}
				}
				if !isLeader {
					log.Println(name, "is not leader")
					continue
				}
			}

			keepAliveRes, err := cli.KeepAliveOnce(ctx, grant.ID)
			if err != nil {
				return err
			}
			log.Println("leader", name, keepAliveRes.TTL)

			id, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			cli.Put(ctx, "/sessions/hoge", id.String())
			cli.Delete(ctx, "/sessions/hoge")
		}
	}
}

func watcher(ctx context.Context) error {
	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		return err
	}
	defer cli.Close()

	watch := cli.Watch(ctx, "/sessions/", clientv3.WithPrefix())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case watchResp, ok := <-watch:
			if !ok {
				return errors.New("watch channel closed")
			}

			if watchResp.Canceled {
				return watchResp.Err()
			}

			for _, event := range watchResp.Events {
				log.Println(event.Type, string(event.Kv.Key), string(event.Kv.Value), event.Kv.Version)
			}
		}
	}
}
