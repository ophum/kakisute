package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ophum/kakisute/sakuracloud-simplemq/simplemq"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func main() {
	mq := simplemq.NewSimpleMQClient(config.QueueName, config.Token)

	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open("database.sqlite"))
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()
	db.AutoMigrate(&Job{})

	e := echo.New()
	e.Use(middleware.Logger(), middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/jobs", func(c echo.Context) error {
		var jobs []*Job
		if err := db.Order("created_at DESC").Find(&jobs).Error; err != nil {
			return err
		}

		return c.JSON(http.StatusOK, jobs)
	})

	e.POST("jobs", func(c echo.Context) error {
		var req struct {
			Name string `json:"name"`
		}
		if err := c.Bind(&req); err != nil {
			return err
		}

		var job Job
		if err := db.Transaction(func(tx *gorm.DB) error {
			id, err := uuid.NewRandom()
			if err != nil {
				return err
			}

			log.Println("create job", id.String())
			encodeID := base64.StdEncoding.EncodeToString([]byte(id.String()))
			res, err := mq.SendMessage(ctx, &simplemq.SendMessageRequest{
				Content: encodeID,
			})
			if err != nil {
				return err
			}

			job = Job{
				ID:        id,
				Name:      req.Name,
				MessageID: res.Message.ID,
				Status:    "Pending",
				CreatedAt: time.Now(),
			}
			if err := db.Create(&job).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, job)
	})

	e.PATCH("/jobs/:id/status", func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return err
		}

		var req struct {
			Status string `json:"status"`
		}
		if err := c.Bind(&req); err != nil {
			return err
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			var job Job
			if err := tx.Where("id = ?", id).First(&job).Error; err != nil {
				return err
			}

			if err := tx.Model(&Job{}).Where("id = ?", id).
				UpdateColumn("status", req.Status).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	})

	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}

}

type Job struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
