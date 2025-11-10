package main

import (
	"errors"
	"io"
	"log/slog"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger(), middleware.Recover())
	e.Static("/", "./public")

	e.GET("/ws", func(c echo.Context) error {
		slog.Info("connect http")
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			slog.Info("upgrade websocket")
			defer func() {
				slog.Info("close websocket")
			}()

			b := make([]byte, 10)
			i := 0
			// control frame: ping
			go func() {
				t := time.NewTicker(time.Second)
				defer t.Stop()
				w, err := ws.NewFrameWriter(websocket.PingFrame)
				if err != nil {
					slog.Error("failed to ws.NewFrameWriter", "error", err)
					return
				}
				defer w.Close()

				for {
					select {
					case <-t.C:
						if _, err := w.Write([]byte("ping")); err != nil {
							slog.Error("failed to Ping", "error", err)
							return
						}
					case <-c.Request().Context().Done():
						slog.Info("ping/pong goroutine done")
						return
					}
				}
			}()
			for {
				n, err := io.ReadFull(ws, b)
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
					slog.Info("EOF")
					break
				}
				if err != nil {
					slog.Error("failed to io.ReadFull", "error", err)
					break
				}
				slog.Info("read", "n", n, "msg", b)

				if err := websocket.Message.Send(ws, "hello "+strconv.Itoa(i)); err != nil {
					slog.Error("failed to send websocket", "error", err)
					break
				}
				i++
			}
		}).ServeHTTP(c.Response(), c.Request())
		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))
}
