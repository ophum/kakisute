package main

import "google.golang.org/protobuf/compiler/protogen"

func main() {
	protogen.Options{}.Run(func(p *protogen.Plugin) error {
		for _, file := range p.Files {
			if !file.Generate {
				continue
			}
			generate(p, file)
		}
		return nil
	})
}

func generate(p *protogen.Plugin, file *protogen.File) {
	filename := file.GeneratedFilenamePrefix + "_gorm_handler.pb.go"
	g := p.NewGeneratedFile(filename, file.GoImportPath)

	g.P("package ", file.GoPackageName)
	g.P("")
	g.P("import (")
	g.P(`"bytes"`)
	g.P(`"encoding/json"`)
	g.P(`"io"`)
	g.P(`"log"`)
	g.P(`"net/http"`)
	g.P(`"strconv"`)
	g.P("")
	g.P(`"gorm.io/gorm"`)
	g.P(")")

	for _, m := range file.Messages {
		g.P(`type `, m.GoIdent, `Model struct {`)
		g.P(`gorm.Model`)
		for _, f := range m.Fields {
			g.P(f.GoName, ` `, f.Desc.Kind())
		}
		g.P(`}`)
		g.P("")
		g.P(`type `, m.GoIdent, `Handler struct {`)
		g.P(`db *gorm.DB`)
		g.P(`*http.ServeMux`)
		g.P(`}`)
		g.P("")
		g.P(`func New`, m.GoIdent, `Handler(db *gorm.DB) (*`, m.GoIdent, `Handler, error){`)
		g.P(`if err := db.AutoMigrate(&`, m.GoIdent, `Model{}); err != nil {`)
		g.P(`return nil, err`)
		g.P(`}`)
		g.P(`handler := &`, m.GoIdent, `Handler{db: db, ServeMux: http.NewServeMux()}`)
		g.P(`handler.HandleFunc("GET /`, m.GoIdent, `/{id}", handler.handleGet)`)
		g.P(`handler.HandleFunc("POST /`, m.GoIdent, `", handler.handleCreate)`)
		g.P(`return handler, nil`)
		g.P(`}`)
		g.P("")
		g.P(`func (h *`, m.GoIdent, `Handler) handleGet(w http.ResponseWriter, r *http.Request) {`)
		g.P(`idStr := r.PathValue("id")`)
		g.P(`id, err := strconv.Atoi(idStr)`)
		g.P(`if err != nil {`)
		g.P(`log.Println(err)`)
		g.P(`w.WriteHeader(http.StatusInternalServerError)`)
		g.P(`return`)
		g.P(`}`)
		g.P(`var v `, m.GoIdent, `Model`)
		g.P(`if err := h.db.Where("id = ?", id).First(&v).Error; err != nil {`)
		g.P(`log.Println(err)`)
		g.P(`w.WriteHeader(http.StatusInternalServerError)`)
		g.P(`return`)
		g.P(`}`)
		g.P(`w.WriteHeader(http.StatusOK)`)
		g.P(`b := bytes.Buffer{}`)
		g.P(`if err := json.NewEncoder(&b).Encode(&v); err != nil {`)
		g.P(`log.Println(err)`)
		g.P(`w.WriteHeader(http.StatusInternalServerError)`)
		g.P(`return`)
		g.P(`}`)
		g.P(`if _, err :=w.Write(b.Bytes()); err != nil {`)
		g.P(`log.Println(err)`)
		g.P(`return`)
		g.P(`}`)
		g.P("}")
		g.P("")
		g.P(`func (h *`, m.GoIdent, `Handler) handleCreate(w http.ResponseWriter, r *http.Request) {`)
		g.P(`var req `, m.GoIdent, "Model")
		g.P(`bodyBytes, err := io.ReadAll(r.Body)`)
		g.P(`r.Body.Close()`)
		g.P(`if err != nil {`)
		g.P(`log.Println("failed to decode request body", err)`)
		g.P(`w.WriteHeader(http.StatusInternalServerError)`)
		g.P(`return`)
		g.P(`}`)
		g.P(`if err := json.Unmarshal(bodyBytes, &req); err != nil {`)
		g.P(`log.Println("failed to decode request body", err)`)
		g.P(`w.WriteHeader(http.StatusInternalServerError)`)
		g.P(`return`)
		g.P(`}`)
		g.P(`if err := h.db.Create(&req).Error; err != nil {`)
		g.P(`log.Println("failed to create record", err)`)
		g.P(`w.WriteHeader(http.StatusInternalServerError)`)
		g.P(`return`)
		g.P(`}`)
		g.P(`w.WriteHeader(http.StatusCreated)`)
		g.P(`b := bytes.Buffer{}`)
		g.P(`if err := json.NewEncoder(&b).Encode(&req); err != nil {`)
		g.P(`log.Println("failedd to encode response body", err)`)
		g.P(`w.WriteHeader(http.StatusInternalServerError)`)
		g.P(`return`)
		g.P(`}`)
		g.P(`if _, err :=w.Write(b.Bytes()); err != nil {`)
		g.P(`log.Println(err)`)
		g.P(`return`)
		g.P(`}`)
		g.P("}")
		g.P("")
	}

}
