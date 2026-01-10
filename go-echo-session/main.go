package main

import (
	"crypto/rand"
	"flag"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
)

var hmacKey = []byte(rand.Text() + rand.Text())
var blockKey = []byte(rand.Text() + rand.Text())[:32]

func main() {
	enableEncryption := flag.Bool("encryption", false, "enable encryption")
	flag.Parse()

	e := echo.New()
	e.Use(middleware.RequestLogger(), middleware.Recover())

	keyPairs := [][]byte{hmacKey}
	if *enableEncryption {
		keyPairs = append(keyPairs, blockKey)
	}
	store := sessions.NewCookieStore(keyPairs...)
	e.Use(session.Middleware(store))

	e.GET("/", func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return err
		}
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
		}
		sess.Values["foo"] = "bar"
		count, _ := sess.Values["count"].(int)
		sess.Values["count"] = count + 1
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]any{
			"message": "ok",
		})
	})
	e.GET("/session", func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			if !err.(securecookie.Error).IsDecode() {
				return errors.Wrap(err, "failed to get session")
			}
			sess.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   3600,
				HttpOnly: true,
			}
			sess.Values = map[any]any{
				"foo":   "",
				"count": 0,
			}
			if err := sess.Save(c.Request(), c.Response()); err != nil {
				return errors.Wrap(err, "failed to save session")
			}
		}
		return c.JSON(http.StatusOK, map[string]any{
			"session": map[string]any{
				"foo":   sess.Values["foo"].(string),
				"count": sess.Values["count"].(int),
			},
		})
	})

	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}
