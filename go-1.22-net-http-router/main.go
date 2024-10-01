package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
)

type Todo struct {
	ID    int
	Title string
	Done  bool
}

var lastID = 0
var todos Todos = []*Todo{}

func newID() int {
	lastID++
	return lastID
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /todos", getTodos)
	mux.HandleFunc("GET /todos/{id}", getTodo)
	mux.HandleFunc("POST /todos", createTodo)
	mux.HandleFunc("POST /todos/{id}/done", doneTodo)
	mux.HandleFunc("DELETE /todos/{id}/done", deleteDoneTodo)

	log.Println("serve :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}

func responseJSON(w http.ResponseWriter, statusCode int, data any) {
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Println(err)
			return
		}
	}
}

type Todos []*Todo

func (t Todos) index(id int) (int, error) {
	i := slices.IndexFunc(t, func(t *Todo) bool {
		return t.ID == id
	})
	if i == -1 {
		return 0, errors.New("not found")
	}
	return i, nil
}

func (t Todos) findByID(id int) (*Todo, error) {
	i, err := t.index(id)
	if err != nil {
		return nil, err
	}
	return &Todo{
		ID:    t[i].ID,
		Title: t[i].Title,
		Done:  t[i].Done,
	}, nil
}

func (t Todos) update(todo *Todo) error {
	i, err := t.index(todo.ID)
	if err != nil {
		return err
	}
	*t[i] = *todo
	return nil
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	ret := make([]*Todo, 0, len(todos))
	for _, todo := range todos {
		ret = append(ret, &Todo{
			ID:    todo.ID,
			Title: todo.Title,
			Done:  todo.Done,
		})
	}

	responseJSON(w, http.StatusOK, ret)
}

type GetTodoRequest struct {
	ID int `uri:"id"`
}

func getTodo(w http.ResponseWriter, r *http.Request) {
	var req GetTodoRequest
	if err := bindURI(r, &req); err != nil {
		responseJSON(w, http.StatusBadRequest, nil)
		return
	}

	todo, err := todos.findByID(req.ID)
	if err != nil {
		responseJSON(w, http.StatusNotFound, nil)
		return
	}

	responseJSON(w, http.StatusOK, todo)
}

type CreateTodoRequest struct {
	Title string
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var req CreateTodoRequest
	if err := bindJSON(r, &req); err != nil {
		responseJSON(w, http.StatusBadRequest, nil)
		return
	}

	todos = append(todos, &Todo{
		ID:    newID(),
		Title: req.Title,
		Done:  false,
	})

	responseJSON(w, http.StatusCreated, nil)
}

type DoneTodoRequest struct {
	ID int `uri:"id"`
}

func doneTodo(w http.ResponseWriter, r *http.Request) {
	var req DoneTodoRequest
	if err := bindURI(r, &req); err != nil {
		responseJSON(w, http.StatusBadRequest, nil)
		return
	}

	todo, err := todos.findByID(req.ID)
	if err != nil {
		responseJSON(w, http.StatusNotFound, nil)
		return
	}

	todo.Done = true

	if err := todos.update(todo); err != nil {
		responseJSON(w, http.StatusNotFound, nil)
		return
	}
	responseJSON(w, http.StatusOK, nil)
}

type DeleteDoneTodoRequest struct {
	ID int `uri:"id"`
}

func deleteDoneTodo(w http.ResponseWriter, r *http.Request) {
	var req DeleteDoneTodoRequest
	if err := bindURI(r, &req); err != nil {
		responseJSON(w, http.StatusBadRequest, nil)
		return
	}

	todo, err := todos.findByID(req.ID)
	if err != nil {
		responseJSON(w, http.StatusNotFound, nil)
		return
	}

	todo.Done = false

	if err := todos.update(todo); err != nil {
		responseJSON(w, http.StatusNotFound, nil)
		return
	}
	responseJSON(w, http.StatusOK, nil)
}
