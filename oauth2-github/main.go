package main

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ClientID     string `yaml:"clientID"`
	ClientSecret string `yaml:"clientSecret"`
	RedirectURL  string `yaml:"redirectURL"`
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

	gob.RegisterName("token", &oauth2.Token{})
	gob.RegisterName("client", &Client{})
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	conf := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes: []string{
			"read:user",
			"user:email",
			"repo",
			"metadata:read",
			"read:org",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://github.com/login/oauth/authorize",
			TokenURL:  "https://github.com/login/oauth/access_token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}

	storeKeyPairs, err := randomString()
	if err != nil {
		return err
	}
	store := memstore.NewStore([]byte(storeKeyPairs))

	r := gin.Default()
	r.Use(sessions.Sessions("sessions", store))

	r.GET("/oauth/github", func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		state, err := randomString()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		session.Set("state", state)
		session.Save()
		log.Println(session.Get("state"))

		url := conf.AuthCodeURL(state)

		ctx.Redirect(http.StatusFound, url)
	})
	r.GET("/oauth/github/callback", func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		sessionState, ok := session.Get("state").(string)
		if !ok {
			ctx.AbortWithStatus(http.StatusPreconditionFailed)
			return
		}

		var callbackParams struct {
			Code  string `form:"code"`
			State string `form:"state"`
		}
		if err := ctx.ShouldBindQuery(&callbackParams); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
		log.Println(callbackParams)
		log.Println(sessionState)

		if callbackParams.State != sessionState {
			ctx.AbortWithStatus(http.StatusPreconditionFailed)
			return
		}

		token, err := conf.Exchange(ctx, callbackParams.Code)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		session.Set("token", token)
		session.Save()
		log.Println("token:", token)
		ctx.Redirect(http.StatusFound, "http://localhost:5173/")
	})

	authed := r.Group("")
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"}
	corsConfig.AllowCredentials = true
	authed.Use(cors.New(corsConfig))
	authed.Use(func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		token, ok := session.Get("token").(*oauth2.Token)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		httpClient := conf.Client(ctx, token)
		client := newClient(httpClient)

		ctx.Set("client", client)
		ctx.Next()
	})
	authed.GET("/api/user", func(ctx *gin.Context) {
		client, err := getClient(ctx)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		var user map[string]any
		if err := client.get("https://api.github.com/user", &user); err != nil {
			log.Println("failed to retrieve user data")
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	})
	authed.GET("/api/user/orgs", func(ctx *gin.Context) {
		client, err := getClient(ctx)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		var orgs []map[string]any
		if err := client.get("https://api.github.com/user/orgs", &orgs); err != nil {
			log.Println("failed to retrieve orgs")
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"orgs": orgs,
		})
	})
	authed.GET("/api/repos", func(ctx *gin.Context) {
		client, err := getClient(ctx)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		page := ctx.Query("page")
		if page == "" {
			page = "1"
		}
		var repos []map[string]any
		if err := client.get("https://api.github.com/user/repos?sort=desc&affiliation=owner&per_page=10&page="+page, &repos); err != nil {
			log.Println("failed to retrieve repo list")
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"repos": repos,
		})
	})
	return r.Run(":8080")
}

type Client struct {
	c *http.Client
}

func newClient(c *http.Client) *Client {
	return &Client{c}
}

func (c *Client) get(url string, v any) error {
	resp, err := c.c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println(resp.Status)
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Println(body)
		return errors.New("failed to get")
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

func randomString() (string, error) {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"

	s := ""
	for i := 0; i < 32; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))-1))
		if err != nil {
			return "", err
		}
		s += string(letters[n.Int64()])
	}
	return s, nil
}

func getClient(ctx *gin.Context) (*Client, error) {
	c, ok := ctx.Get("client")
	if !ok {
		return nil, errors.New("failed to get client")
	}
	client, ok := c.(*Client)
	if !ok {
		return nil, errors.New("invalid type client")
	}
	return client, nil
}
