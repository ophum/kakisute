package user

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

type UserModel struct {
	gorm.Model
	Name        string
	Description string
	Age         uint32
}

type UserHandler struct {
	db *gorm.DB
	*http.ServeMux
}

func NewUserHandler(db *gorm.DB) (*UserHandler, error) {
	if err := db.AutoMigrate(&UserModel{}); err != nil {
		return nil, err
	}
	handler := &UserHandler{db: db, ServeMux: http.NewServeMux()}
	handler.HandleFunc("GET /User/{id}", handler.handleGet)
	handler.HandleFunc("POST /User", handler.handleCreate)
	return handler, nil
}

func (h *UserHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var v UserModel
	if err := h.db.Where("id = ?", id).First(&v).Error; err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(&v); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(b.Bytes()); err != nil {
		log.Println(err)
		return
	}
}

func (h *UserHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req UserModel
	bodyBytes, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Println("failed to decode request body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Println("failed to decode request body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := h.db.Create(&req).Error; err != nil {
		log.Println("failed to create record", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(&req); err != nil {
		log.Println("failedd to encode response body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(b.Bytes()); err != nil {
		log.Println(err)
		return
	}
}
