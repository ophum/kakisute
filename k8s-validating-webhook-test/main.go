package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

var allowSuffix string

func main() {
	flag.StringVar(&allowSuffix, "allow-suffix", "", "")
	flag.Parse()

	http.HandleFunc("/validate", handleValidate)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func reqID() string {
	id := make([]byte, 4)
	_, _ = rand.Read(id)
	return hex.EncodeToString(id)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	id := reqID()
	status := 200
	defer func() {
		slog.Info("/validate", "id", id, "status", status)
	}()
	arReq := v1beta1.AdmissionReview{}
	if err := json.NewDecoder(r.Body).Decode(&arReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		status = http.StatusBadRequest
		return
	}

	b, _ := json.MarshalIndent(arReq, "", " ")
	fmt.Println(string(b))

	if arReq.Request == nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		status = http.StatusBadRequest
		return
	}

	var httpRoute v1.HTTPRoute
	if err := json.Unmarshal(arReq.Request.Object.Raw, &httpRoute); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		status = http.StatusBadRequest
		return
	}

	res := v1beta1.AdmissionReview{
		TypeMeta: arReq.TypeMeta,
		Response: &v1beta1.AdmissionResponse{
			UID:     arReq.Request.UID,
			Allowed: true,
		},
	}
	if len(httpRoute.Spec.Hostnames) == 0 {
		res.Response.Allowed = false
		res.Response.Result = &metav1.Status{
			Message: "invalid .spec.hostnames: at least one hostname must be specified",
		}
	}
	for i, hostname := range httpRoute.Spec.Hostnames {
		if !strings.HasSuffix(string(hostname), allowSuffix) {
			res.Response.Allowed = false
			res.Response.Result = &metav1.Status{
				Message: fmt.Sprintf("invalid .spec.hostnames[%d]: %s is not allowed", i, string(hostname)),
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		slog.Error("failed to encode and write response body", "error", err)
	}
}
