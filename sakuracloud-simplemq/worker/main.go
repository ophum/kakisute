package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sync"
	"time"

	"github.com/bit101/go-ansi"
	"github.com/ophum/kakisute/sakuracloud-simplemq/simplemq"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Token     string `yaml:"token"`
	QueueName string `yaml:"queueName"`
}

var config Config

func init() {
	f, err := os.Open("config.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}
}

type WorkerStatus struct {
	MessageID string
	StartedAt time.Time
	Status    int
}

var workerStatus = map[string]*WorkerStatus{}
var mu sync.RWMutex

func updateWorkerStatus(id string, status int) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := workerStatus[id]; !ok {
		workerStatus[id] = &WorkerStatus{
			MessageID: id,
			StartedAt: time.Now(),
		}
	}
	workerStatus[id].Status = status
}
func main() {
	c := simplemq.NewSimpleMQClient(config.QueueName, config.Token)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	autoUpdater := &AutoUpdater{
		msgs:     map[string]time.Time{},
		extended: map[string]struct{}{},
		mu:       sync.RWMutex{},
	}
	go autoUpdater.Start(ctx, c, time.Second*30)

	go func() {
		tick := time.NewTicker(time.Millisecond * 800)
		defer tick.Stop()

		ansi.ClearScreen()
		for {
			select {
			case <-tick.C:
				ansi.MoveTo(0, 0)
				ansi.Printf(ansi.Red, "worker process\n")
				ansi.Println(ansi.BoldWhite, "started_at                message_id                           status")
				mu.RLock()
				out := []*WorkerStatus{}
				for _, status := range workerStatus {
					out = append(out, status)
				}
				slices.SortStableFunc(out, func(a, b *WorkerStatus) int {
					return cmp.Compare(a.StartedAt.Unix(), b.StartedAt.Unix())
				})
				for _, o := range out {
					ansi.Printf(ansi.Blue, "%s %s %d\n", o.StartedAt.Format(time.RFC3339), o.MessageID, o.Status)
				}
				mu.RUnlock()
			case <-ctx.Done():
				return
			}
		}
	}()
	sem := semaphore.NewWeighted(2)
	for {
		res, err := c.ReceiveMessage(ctx)
		if err != nil {
			log.Fatal(err)
		}

		if len(res.Messages) == 0 {
			time.Sleep(time.Second)
			continue
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			log.Fatal(err)
		}
		msg := res.Messages[0]
		if err := updateJobStatus(ctx, msg, "Running"); err != nil {
			log.Fatal(err)
		}
		autoUpdater.Register(msg.ID, msg.CreatedAt)
		go worker(ctx, msg, func(err error) {
			defer func() {
				autoUpdater.Unregister(msg.ID)
				sem.Release(1)
			}()
			if err := c.AckMessage(ctx, msg.ID); err != nil {
				if !errors.Is(err, simplemq.ErrNotFound) {
					log.Fatal(err)
				}
			}
			updateJobStatus(ctx, msg, "Done")
			autoUpdater.Unregister(msg.ID)
		})
	}
}

func worker(ctx context.Context, msg *simplemq.Message, doneFn func(err error)) {
	updateWorkerStatus(msg.ID, 0)
	for i := 0; i <= 100; i++ {
		updateWorkerStatus(msg.ID, i)
		time.Sleep(time.Millisecond * 100)
	}
	doneFn(nil)
}

func updateJobStatus(ctx context.Context, msg *simplemq.Message, status string) error {
	id, err := base64.StdEncoding.DecodeString(msg.Content)
	if err != nil {
		return err
	}

	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(&struct {
		Status string `json:"status"`
	}{
		Status: status,
	}); err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("http://localhost:8080/jobs/%s/status", string(id)),
		&b,
	)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "failed to update")
	}
	defer httpRes.Body.Close()

	log.Println(httpRes)
	return nil
}

type AutoUpdater struct {
	msgs     map[string]time.Time // id: created_at
	extended map[string]struct{}
	mu       sync.RWMutex
}

func (a *AutoUpdater) Register(id string, createdAt time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.msgs[id]; ok {
		return
	}
	a.msgs[id] = createdAt
}

func (a *AutoUpdater) Unregister(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.msgs, id)
}

func (a *AutoUpdater) Start(ctx context.Context, c simplemq.SimpleMQClient, timeout time.Duration) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case t := <-tick.C:
			a.mu.RLock()
			for id, createdAt := range a.msgs {
				if t.Sub(createdAt)%timeout < timeout-time.Second*10 {
					if _, ok := a.extended[id]; ok {
						delete(a.extended, id)
						continue
					}
					continue
				}
				if _, ok := a.extended[id]; ok {
					continue
				}
				if err := c.UpdateMessageTimeout(ctx, id); err != nil {
					if errors.Is(err, simplemq.ErrNotFound) {
						continue
					}
					log.Println(err)
					continue
				}
				a.extended[id] = struct{}{}
			}
			a.mu.RUnlock()
		case <-ctx.Done():
			return
		}
	}
}
