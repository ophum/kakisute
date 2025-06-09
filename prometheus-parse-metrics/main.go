package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

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

	outMetrics := []*io_prometheus_client.MetricFamily{}
	for _, metric := range metrics {
		outMetrics = append(outMetrics, metric)
		for _, v := range metric.GetMetric() {
			name := "greeting"
			value := "hello ^^"
			v.Label = append(v.Label, &io_prometheus_client.LabelPair{
				Name:  &name,
				Value: &value,
			})
		}
	}

	b := bytes.Buffer{}
	for _, v := range metrics {
		_, err := expfmt.MetricFamilyToText(&b, v)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println(b.String())
}
