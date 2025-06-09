package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"slices"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func main() {
	target := os.Args[1]

	res, err := http.Get(target)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	parser := expfmt.TextParser{}

	metrics, err := parser.TextToMetricFamilies(res.Body)
	if err != nil {
		panic(err)
	}

	names := []string{}
	for _, metric := range metrics {
		names = append(names, metric.GetName())
		for _, v := range metric.GetMetric() {
			name := "greeting"
			value := "hello ^^"
			v.Label = append(v.Label, &io_prometheus_client.LabelPair{
				Name:  &name,
				Value: &value,
			})
		}
	}

	slices.Sort(names)
	result := []*io_prometheus_client.MetricFamily{}
	for _, name := range names {
		result = append(result, metrics[name])
	}

	b := bytes.Buffer{}
	for _, v := range result {
		_, err := expfmt.MetricFamilyToText(&b, v)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println(b.String())
}
