package main

import (
	"context"
	controllertest "controller-test"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := worker(ctx); err != nil {
		slog.Error("failed to run worker", "error", err)
		os.Exit(1)
	}
}

func getSnapshot(ctx context.Context) (*controllertest.Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/snapshot", nil)
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to do request", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("unexpected status code", "status", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var snapshot controllertest.Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		slog.Error("failed to decode response", "error", err)
		return nil, err
	}
	return &snapshot, nil
}

func listEvent(ctx context.Context, cursor int) ([]*controllertest.Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/events", nil)
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return nil, err
	}

	v := req.URL.Query()
	v.Set("cursor", fmt.Sprint(cursor))
	req.URL.RawQuery = v.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to do request", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("unexpected status code", "status", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var events []*controllertest.Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		slog.Error("failed to decode response", "error", err)
		return nil, err
	}
	return events, nil
}
func worker(ctx context.Context) error {
	timer := time.NewTicker(time.Second)
	defer timer.Stop()

	snapshot, err := getSnapshot(ctx)
	if err != nil {
		return err
	}

	users, cursor := snapshot.Users, snapshot.Cursor

	displayUsers(users)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			events, err := listEvent(ctx, cursor)
			if err != nil {
				slog.Error("failed to list events", "error", err)
				continue
			}
			if len(events) == 0 {
				slog.Info("no new events")
				continue
			}
			displayEvents(events)
			users, cursor, err = applyEvents(users, cursor, events)
			if err != nil {
				slog.Error("failed to process event", "error", err)
			}
			displayUsers(users)
		}
	}
}

func applyEvents(users []*controllertest.User, cursor int, events []*controllertest.Event) ([]*controllertest.User, int, error) {
	for _, event := range events {
		var err error
		users, err = applyEvent(users, event)
		if err != nil {
			return users, cursor, fmt.Errorf("event %d: %w", event.ID, err)
		}
		cursor = event.ID
	}
	return users, cursor, nil
}

func applyEvent(users []*controllertest.User, event *controllertest.Event) ([]*controllertest.User, error) {
	switch event.Action {
	case "create_user", "update_user":
		var user controllertest.User
		if err := json.Unmarshal([]byte(event.Body), &user); err != nil {
			return users, err
		}
		slog.Info("process", "id", event.ID, "action", event.Action)
		for i, existing := range users {
			if existing.ID == user.ID {
				users[i] = &user
				return users, nil
			}
		}
		return append(users, &user), nil
	default:
		slog.Info("skip unknown event", "id", event.ID, "action", event.Action)
		return users, nil
	}
}

func displayUsers(users []*controllertest.User) {
	fmt.Println("----- Users -----")
	for _, user := range users {
		fmt.Printf("user id: %d, name: %s\n", user.ID, user.Name)
	}
	fmt.Println()
}

func displayEvents(events []*controllertest.Event) {
	fmt.Println("----- Events -----")
	for _, event := range events {
		fmt.Printf("event id: %d, action: %s, body: %s, event at %s\n", event.ID, event.Action, event.Body, event.EventAt)
	}
	fmt.Println()
}
