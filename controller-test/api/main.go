package main

import (
	"context"
	controllertest "controller-test"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "modernc.org/sqlite"

	echo "github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := api(ctx); err != nil {
		slog.Error("failed to run api", "error", err)
		os.Exit(1)
	}
}

var db *sql.DB

func api(ctx context.Context) error {
	var err error
	db, err = sql.Open("sqlite", "file:test.db?cache=shared&mode=rwc")
	if err != nil {
		return err
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		return err
	}

	e := newRouter()
	sc := echo.StartConfig{
		Address: ":8080",
	}
	return sc.Start(ctx, e)
}

func initDB(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT
	)`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	event_at DATETIME,
	action TEXT,
	body TEXT
	)`); err != nil {
		return err
	}

	return nil
}

func newRouter() *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestLogger(), middleware.Recover())
	e.GET("/list", list)
	e.GET("/snapshot", snapshot)
	e.POST("/create", create)
	e.PUT("/put", put)
	e.GET("/events", listEvent)

	return e
}

func snapshot(c *echo.Context) error {
	ctx := c.Request().Context()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var cursor int
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(id), 0) FROM events").Scan(&cursor); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, "SELECT id, name FROM users")
	if err != nil {
		return err
	}
	users := make([]*controllertest.User, 0)
	for rows.Next() {
		var user controllertest.User
		if err := rows.Scan(&user.ID, &user.Name); err != nil {
			rows.Close()
			return err
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(200, struct {
		Users  []*controllertest.User `json:"users"`
		Cursor int                    `json:"cursor"`
	}{users, cursor})
}

func list(c *echo.Context) error {
	rows, err := db.Query("SELECT id, name FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return rows.Err()
	}

	var users []*controllertest.User
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		users = append(users, &controllertest.User{
			ID:   id,
			Name: name,
		})
	}
	return c.JSON(200, users)
}

func create(c *echo.Context) error {
	var user controllertest.User
	if err := c.Bind(&user); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec("INSERT INTO users (name) VALUES (?)", user.Name)
	if err != nil {
		tx.Rollback()
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}
	user.ID = int(id)

	b, err := json.Marshal(user)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("INSERT INTO events (event_at, action, body) VALUES (datetime('now'), ?, ?)", "create_user", b)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(201, user)
}

func put(c *echo.Context) error {
	var user controllertest.User
	if err := c.Bind(&user); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE users SET name = ? WHERE id = ?", user.Name, user.ID)
	if err != nil {
		tx.Rollback()
		return err
	}
	b, err := json.Marshal(user)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("INSERT INTO events (event_at, action, body) VALUES (datetime('now'), ?, ?)", "update_user", b)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(200, user)
}

func listEvent(c *echo.Context) error {
	cursor := c.QueryParam("cursor")
	id, _ := strconv.Atoi(cursor)
	rows, err := db.Query("SELECT id, event_at, action, body FROM events WHERE id > ? ORDER BY id ASC", id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return rows.Err()
	}

	var events []*controllertest.Event
	for rows.Next() {
		var id int
		var eventAt, action, body string
		if err := rows.Scan(&id, &eventAt, &action, &body); err != nil {
			return err
		}
		events = append(events, &controllertest.Event{
			ID:      id,
			EventAt: eventAt,
			Action:  action,
			Body:    body,
		})
	}
	return c.JSON(200, events)
}
