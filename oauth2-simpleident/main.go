package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func main() {
	conf := &oauth2.Config{
		ClientID:     "01923bae-0a3c-7e79-9f1c-41bd843c3281",
		ClientSecret: "x003tyCYZE9YXNoB2Pq3FdOIN4ky1thZSfRJYcrjnDaNDz3LowcgyysXUSxAuW66",
		RedirectURL:  "http://localhost:8081/oauth/callback",
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:8080/oauth2/authorize",
			TokenURL: "http://localhost:8080/oauth2/token",
		},
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	store := memstore.NewStore([]byte("keypair..."))
	r.Use(sessions.Sessions("kakisute-oauth2-simpleident", store))

	r.GET("/", func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		flashMessage, _ := session.Get("flash_message").(string)
		ctx.HTML(http.StatusOK, "index.html.tmpl", gin.H{
			"FlashMessage": flashMessage,
		})
	})
	r.GET("/authed", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID, ok := session.Get("user_id").(string)
		if !ok {
			session.Set("flash_message", "ログインしてください")
			session.Save()
			ctx.Redirect(http.StatusFound, "/")
			return
		}
		username, ok := session.Get("username").(string)
		if !ok {
			session.Set("flash_message", "ログインしてください")
			session.Save()
			return
		}

		ctx.HTML(http.StatusOK, "authed.html.tmpl", gin.H{
			"ID":       userID,
			"Username": username,
		})
	})

	r.GET("/oauth", func(ctx *gin.Context) {
		state, err := uuid.NewRandom()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		session := sessions.Default(ctx)
		session.Set("state", state.String())
		if err := session.Save(); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		url := conf.AuthCodeURL(state.String())
		ctx.Redirect(http.StatusFound, url)
	})

	r.GET("/oauth/callback", func(ctx *gin.Context) {
		var callbackParams struct {
			Code  string `form:"code"`
			State string `form:"state"`
		}
		if err := ctx.ShouldBindQuery(&callbackParams); err != nil {
			ctx.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}

		session := sessions.Default(ctx)
		state, ok := session.Get("state").(string)
		if !ok {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if callbackParams.State != state {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		token, err := conf.Exchange(ctx, callbackParams.Code)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		client := conf.Client(ctx, token)

		res, err := client.Get("http://localhost:8080/api/userinfo")
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer res.Body.Close()

		var v map[string]any
		if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		userID := v["id"]
		username := v["username"]

		session.Set("user_id", userID)
		session.Set("username", username)

		if err := session.Save(); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Redirect(http.StatusFound, "/authed")
	})

	if err := r.Run(":8081"); err != nil {
		panic(err)
	}
}
