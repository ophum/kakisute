package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var config Config

type Config struct {
	Address  string
	Port     int
	Database string
	Username string
	Password string
}

func readConfig() {
	f, err := os.Open("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		log.Fatal(err)
	}
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "memoを追加します。",
				Action: func(ctx *cli.Context) error {
					readConfig()
					db, err := getDB()
					if err != nil {
						return err
					}
					memo := Memo{
						Content: ctx.Args().First(),
					}
					if err := db.Create(&memo).Error; err != nil {
						return err
					}

					log.Printf("created: %d, %s", memo.ID, memo.Content)
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "memoの一覧を表示します。",
				Action: func(ctx *cli.Context) error {
					readConfig()
					db, err := getDB()
					if err != nil {
						return err
					}

					var memos []*Memo
					if err := db.Find(&memos).Error; err != nil {
						return err
					}

					for _, m := range memos {
						log.Printf("id=%d content=%s created_at=%s", m.ID, m.Content, m.CreatedAt.String())
					}
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type Memo struct {
	*gorm.Model
	Content string
}

func getDB() (*gorm.DB, error) {
	//dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=verify-full sslrootcert=ca.crt sslkey=client.roach.key sslcert=client.roach.crt TimeZone=Asia/Tokyo",
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d  TimeZone=Asia/Tokyo",
		config.Address,
		config.Username,
		config.Password,
		config.Database,
		config.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Memo{}); err != nil {
		return nil, err
	}

	db = db.Debug()

	return db, nil
}
