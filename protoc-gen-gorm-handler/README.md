# protoc-gen-gorm-handler

protoc の plugin をお試し実装。
protoc で message を元に GORM の Model と HTTP Handler を生成してみた。

```
$ go build
$ protoc -I. --plugin=./protoc-gen-gorm-handler --gorm-handler_out=. test.proto
```

`pb/user/test_gorm_handler.pb.go`が生成された。

```go
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


```

## example

生成した Handler を使ってみる

```
$ go run example/main.go
```

`/User`というエンドポイントに POST するとデータを DB に登録できる。

```
$ curl -X POST localhost:8080/User -H "Content-Type: application/json" -d '{"Name": "test user", "Description": "hello world", "Age": 30}' | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   244  100   182  100    62  42583  14506 --:--:-- --:--:-- --:--:-- 61000
{
  "ID": 1,
  "CreatedAt": "2025-05-10T01:18:48.615686726+09:00",
  "UpdatedAt": "2025-05-10T01:18:48.615686726+09:00",
  "DeletedAt": null,
  "Name": "test user",
  "Description": "hello world",
  "Age": 30
}
```

`/User/{id}`というエンドポイントに GET すると`{id}`のデータを取得できる。

```
$ curl -X GET localhost:8080/User/1 | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   182  100   182    0     0   196k      0 --:--:-- --:--:-- --:--:--  177k
{
  "ID": 1,
  "CreatedAt": "2025-05-10T01:18:48.615686726+09:00",
  "UpdatedAt": "2025-05-10T01:18:48.615686726+09:00",
  "DeletedAt": null,
  "Name": "test user",
  "Description": "hello world",
  "Age": 30
}
```

sqlite3 で gorm.db の中にデータが入っていることを確認。

```
$ sqlite3 gorm.db
SQLite version 3.45.1 2024-01-30 16:01:20
Enter ".help" for usage hints.
sqlite> select * from user_models;
1|2025-05-10 01:18:48.615686726+09:00|2025-05-10 01:18:48.615686726+09:00||test user|hello world|30
```
