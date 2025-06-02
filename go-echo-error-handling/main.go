package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		er, ok := err.(*Error)
		if ok {
			e.DefaultHTTPErrorHandler(echo.NewHTTPError(
				er.Status,
				map[string]string{
					"code":    er.Code,
					"message": er.Message,
				},
			), c)
			return
		}

		e.DefaultHTTPErrorHandler(err, c)
	}
	e.Use(middleware.Logger(), middleware.Recover())
	e.GET("/", func(c echo.Context) error {
		return NewError(http.StatusUnauthorized, "invalid_token", "invalid token", errors.New("invalid token internal error"))
	})

	e.Start(":8080")
}

type Error struct {
	Status   int
	Code     string
	Message  string
	Internal error
}

func (e Error) Error() string {
	return fmt.Sprintf("status=%d code=%s err=%+v", e.Status, e.Code, e.Internal)
}

func NewError(status int, code, message string, err error) error {
	return &Error{
		Status:   status,
		Code:     code,
		Message:  message,
		Internal: err,
	}
}
