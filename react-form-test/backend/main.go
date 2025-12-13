package main

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

const sessionName = "react-form-test-session"

var baseURL = &url.URL{
	Scheme: "http",
	Host:   "localhost:5173",
}

func frontendURL(path string, query url.Values) *url.URL {
	u := *baseURL
	u.Path = path
	u.RawQuery = query.Encode()
	return &u

}

func run() error {
	e := echo.New()
	e.Use(middleware.RequestLogger(), middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("changeme-secret"))))
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:_csrf",
	}))

	e.GET("/login", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	e.POST("/login", func(c echo.Context) error {
		type LoginForm struct {
			Username string `form:"username"`
			Password string `form:"password"`
		}

		var form LoginForm
		if err := c.Bind(&form); err != nil {
			return err
		}

		slog.Info("req", "username", form.Username, "password", form.Password)

		if form.Username != "user" {
			v := url.Values{}
			v.Set("error", "invalid username or password")
			slog.Info("invalid username")
			return c.Redirect(http.StatusFound, frontendURL("login", v).String())
		}

		if subtle.ConstantTimeCompare([]byte(form.Password), []byte("password")) != 1 {
			v := url.Values{}
			v.Set("error", "invalid username or password")
			slog.Info("invalid passwrod")
			return c.Redirect(http.StatusFound, frontendURL("login", v).String())
		}

		sess, err := session.Get(sessionName, c)
		if err != nil {
			return err
		}

		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
		}
		sess.Values["username"] = form.Username
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, frontendURL("/", url.Values{}).String())
	})

	e.GET("/api/me", func(c echo.Context) error {
		sess, err := session.Get(sessionName, c)
		if err != nil {
			return err
		}

		username, ok := sess.Values["username"].(string)
		if !ok {
			return echo.ErrUnauthorized
		}

		return c.JSON(http.StatusOK, map[string]any{
			"username": username,
		})
	})

	e.POST("/change-username", func(c echo.Context) error {
		type ChangeUsernameForm struct {
			Username string `form:"username"`
		}
		var req ChangeUsernameForm
		if err := c.Bind(&req); err != nil {
			return err
		}

		sess, err := session.Get(sessionName, c)
		if err != nil {
			return err
		}

		sess.Values["username"] = req.Username
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})
	return e.Start(":8080")
}
