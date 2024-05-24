package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ClientID     string `yaml:"clientID"`
	ClientSecret string `yaml:"clientSecret"`
	RedirectURL  string `yaml:"redirectURL"`
	ServerID     string `yaml:"serverID"`
}

var (
	configFile string
	config     Config
)

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "config.yaml")
	flag.Parse()

	f, err := os.Open(configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		log.Fatal(err)
	}
}

func main() {
	conf := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes: []string{
			"identify",
			"email",
			"guilds.members.read",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}

	r := gin.Default()
	r.GET("/oauth/discord", func(ctx *gin.Context) {
		url := conf.AuthCodeURL("statestate")
		ctx.Redirect(http.StatusFound, url)
	})

	r.GET("/oauth/discord/callback", func(ctx *gin.Context) {
		var callbackParams struct {
			Code  string `form:"code"`
			State string `form:"state"`
		}
		if err := ctx.ShouldBindQuery(&callbackParams); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		token, err := conf.Exchange(ctx, callbackParams.Code)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		client := conf.Client(ctx, token)

		if err := authed(ctx, client); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func authed(ctx *gin.Context, client *http.Client) error {
	userRes, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return err
	}
	defer userRes.Body.Close()

	var userResult map[string]any
	if err := json.NewDecoder(userRes.Body).Decode(&userResult); err != nil {
		return err
	}

	guildMemberRes, err := client.Get(fmt.Sprintf("https://discord.com/api/users/@me/guilds/%s/member", config.ServerID))
	if err != nil {
		return err
	}
	defer guildMemberRes.Body.Close()

	var guildMemberResult map[string]any
	if err := json.NewDecoder(guildMemberRes.Body).Decode(&guildMemberResult); err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, gin.H{
		"userStatus":        userRes.StatusCode,
		"user":              userResult,
		"guildMemberStatus": guildMemberRes.StatusCode,
		"guildMember":       guildMemberResult,
	})
	return nil
}
