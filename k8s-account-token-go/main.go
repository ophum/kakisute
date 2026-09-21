package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	kclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	if os.Args[1] == "api1" {
		return api1(kclient)
	}
	return api2(kclient)
}

func api1(kclient kubernetes.Interface) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token, err := kclient.CoreV1().ServiceAccounts("api1").CreateToken(context.TODO(), "default", &authv1.TokenRequest{
			Spec: authv1.TokenRequestSpec{
				Audiences: []string{
					"api2",
				},
				ExpirationSeconds: new(int64(600)),
			},
		}, v1.CreateOptions{})
		if err != nil {
			fmt.Fprintf(w, "error: %+v", err)
			return
		}

		req, err := http.NewRequest(http.MethodGet, "http://api2.api2.svc.cluster.local:8000", nil)
		if err != nil {
			fmt.Fprintf(w, "error: %+v", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.Status.Token)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprintf(w, "error: %+v", err)
			return
		}
		defer res.Body.Close()

		io.Copy(w, res.Body)
	})
	return http.ListenAndServe(":8000", nil)
}

func api2(kclient kubernetes.Interface) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		review, err := kclient.AuthenticationV1().TokenReviews().Create(context.TODO(), &authv1.TokenReview{
			Spec: authv1.TokenReviewSpec{
				Audiences: []string{
					"api2",
				},
				Token: token,
			},
		}, v1.CreateOptions{})
		if err != nil {
			fmt.Fprintf(w, "error: %+v", err)
			return
		}

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(review)
	})
	return http.ListenAndServe(":8000", nil)
}
