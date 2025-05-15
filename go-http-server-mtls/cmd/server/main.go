package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	})
	e.GET("/ping", func(c echo.Context) error {
		commonName := c.Request().TLS.PeerCertificates[0].Subject.CommonName
		return c.JSON(http.StatusOK, map[string]any{
			"message":    "pong",
			"commonName": commonName,
		})
	})

	rootCA, err := os.ReadFile("ca.crt")
	if err != nil {
		return err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(rootCA) {
		return errors.New("invalid ca")
	}

	s := http.Server{
		Addr:    ":8080",
		Handler: e,
		TLSConfig: &tls.Config{
			ClientCAs:  certPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}
	return s.ListenAndServeTLS("server.crt", "server.key")
}
