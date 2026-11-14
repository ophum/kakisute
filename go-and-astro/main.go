package main

import (
	"context"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"

	echo "github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}

//go:embed frontend/dist/*
var frontendDistFS embed.FS

func loadTemplates(paths []string) (*template.Template, error) {
	files, err := fs.Sub(frontendDistFS, "frontend/dist")
	if err != nil {
		return nil, err
	}
	tmpl := template.New("")
	for _, path := range paths {
		f, err := files.Open(path)
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		if _, err := tmpl.New(path).Parse(string(content)); err != nil {

			return nil, err
		}
	}
	return tmpl, nil
}

var mu sync.Mutex
var data = []string{}

func walkTemplates(name string) ([]string, error) {
	entries, err := frontendDistFS.ReadDir(name)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			subNames, err := walkTemplates(name + "/" + entry.Name())
			if err != nil {
				return nil, err
			}
			for _, subName := range subNames {
				names = append(names, entry.Name()+"/"+subName)
			}
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}
		names = append(names, entry.Name())
	}
	return names, nil
}
func run(ctx context.Context) error {
	names, err := walkTemplates("frontend/dist")
	if err != nil {
		return err
	}
	slog.Info("template names", "names", names)
	e := echo.New()
	tmpl, err := loadTemplates(names)
	if err != nil {
		return err
	}
	e.Renderer = &echo.TemplateRenderer{
		Template: tmpl,
	}
	e.StaticFS("_astro/", echo.MustSubFS(frontendDistFS, "frontend/dist/_astro/"))
	e.Use(middleware.RequestLogger(), middleware.Recover())
	e.Use(middleware.CSRF())

	e.GET("/", func(c *echo.Context) error {
		mu.Lock()
		names := slices.Clone(data)
		mu.Unlock()
		posts := []string{}
		for _, t := range tmpl.Templates() {
			if strings.HasPrefix(t.Name(), "posts/") {
				post, ok := strings.CutSuffix(t.Name(), "/index.html")
				if !ok {
					continue
				}
				posts = append(posts, post)
			}
		}
		return c.Render(200, "index.html", map[string]any{
			"Names": names,
			"Posts": posts,
		})
	})
	e.GET("/form", func(c *echo.Context) error {
		return c.Render(200, "form/index.html", map[string]any{
			"Name": "Hello, World",
		})
	})
	e.POST("/post", func(c *echo.Context) error {
		name := c.FormValue("name")
		mu.Lock()
		data = append(data, name)
		mu.Unlock()
		return c.Redirect(http.StatusFound, "/")
	})
	e.GET("/posts/:name", func(c *echo.Context) error {
		name := c.Param("name")
		if len(name) > 8 {
			return echo.ErrNotFound
		}
		for _, t := range tmpl.Templates() {
			if t.Name() == "posts/"+name+"/index.html" {
				return c.Render(200, "posts/"+name+"/index.html", map[string]any{
					"Name": name,
				})
			}
		}
		return echo.ErrNotFound
	})
	sc := echo.StartConfig{
		Address: ":8080",
	}
	return sc.Start(ctx, e)
}
