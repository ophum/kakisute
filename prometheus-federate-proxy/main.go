package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	target := flag.String("target", "http://prometheus:9090/federate", "federate target")
	instancesStr := flag.String("instances", "", "comma separated key value (key1=value1,key2=value2)")
	flag.Parse()

	instances := map[string]string{}
	kvs := strings.Split(*instancesStr, ",")
	for _, kv := range kvs {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			log.Println("invalid key=value,", kv)
			continue
		}
		log.Println(k, v)

		instances[k] = v
	}
	promURL, err := url.Parse(*target)
	if err != nil {
		panic(err)
	}
	log.Println(promURL.String())
	http.HandleFunc("GET /{tenant}/metrics", func(w http.ResponseWriter, r *http.Request) {
		tenant := r.PathValue("tenant")

		instanceName, ok := instances[tenant]
		if !ok {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		u := *promURL

		q := u.Query()
		q.Add("match[]", fmt.Sprintf(`{instance="%s"}`, instanceName))
		u.RawQuery = q.Encode()
		log.Println(u.String())

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			panic(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(body[:20]))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(body); err != nil {
			panic(err)
		}
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
