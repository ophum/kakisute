package main

import (
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type stateStore struct {
	store       map[string]string
	storeLocker map[string]string
	sync.Mutex
}

func main() {
	store := stateStore{
		store:       map[string]string{},
		storeLocker: map[string]string{},
		Mutex:       sync.Mutex{},
	}
	store.store["hoge"] = ""

	r := gin.Default()

	r.GET("/:state", func(ctx *gin.Context) {
		stateName := ctx.Param("state")
		if stateName == "" {
			ctx.Status(http.StatusBadRequest)
			return
		}

		store.Lock()
		defer store.Unlock()

		state, ok := store.store[stateName]
		if !ok {

			ctx.Status(http.StatusNotFound)
		}

		ctx.Data(http.StatusOK, "application/json", []byte(state))
	})
	r.POST("/:state", func(ctx *gin.Context) {
		stateName := ctx.Param("state")
		if stateName == "" {
			ctx.Status(http.StatusBadRequest)
			return
		}
		id := ctx.Query("ID")

		store.Lock()
		defer store.Unlock()

		lockedID, locked := store.storeLocker[stateName]
		if locked && id != lockedID {
			ctx.Status(http.StatusLocked)
			return
		}
		state, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		store.store[stateName] = string(state)
		log.Println("state", string(state))

		ctx.Status(http.StatusOK)
	})
	r.Handle("LOCK", "/:state", func(ctx *gin.Context) {
		stateName := ctx.Param("state")
		if stateName == "" {
			ctx.Status(http.StatusBadRequest)
			return
		}

		var req struct {
			ID string `json:"ID"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		store.Lock()
		defer store.Unlock()

		_, locked := store.storeLocker[stateName]
		if locked {
			ctx.Status(http.StatusLocked)
			return
		}

		store.storeLocker[stateName] = req.ID

		ctx.Status(http.StatusOK)
	})
	r.Handle("UNLOCK", "/:state", func(ctx *gin.Context) {
		stateName := ctx.Param("state")
		if stateName == "" {
			ctx.Status(http.StatusBadRequest)
			return
		}
		var req struct {
			ID string `json:"ID"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		store.Lock()
		defer store.Unlock()

		id, locked := store.storeLocker[stateName]
		if locked && id == req.ID {
			delete(store.storeLocker, stateName)
			ctx.Status(http.StatusOK)
			return
		}

		ctx.Status(http.StatusLocked)
	})
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
