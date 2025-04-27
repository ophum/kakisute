package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	sakurarslog2parquet "github.com/ophum/sakura-rs-log2parquet"
	"gopkg.in/yaml.v3"
)

type Config struct {
	S3 struct {
		AccessKeyID     string `yaml:"accessKeyID"`
		SecretAccessKey string `yaml:"secretAccessKey"`
		Region          string `yaml:"region"`
		Endpoint        string `yaml:"endpoint"`
	} `yaml:"s3"`
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
func run() error {
	dStr := flag.String("partition-duration", "1m", "partition")
	prefix := flag.String("prefix", "test.log.", "prefix")
	fileSize := flag.Int("file-size", 20, "file size (MiB)")
	flag.Parse()
	d, err := time.ParseDuration(*dStr)
	if err != nil {
		return err
	}

	conf, err := loadConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	s3Client := newS3(conf)

	tzJST, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return err
	}

	hosts := []string{}
	for range 10 {
		hosts = append(hosts, uuid.Must(uuid.NewRandom()).String())
	}
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	start := time.Date(2025, 4, 1, 0, 0, 0, 0, tzJST)
	end := start.Add(time.Hour * 24)
	totalBytes := 0
	for t := start; t.Compare(end) < 0; t = t.Add(d) {
		b := bytes.Buffer{}
		for b.Len() < 1024*1024*(*fileSize) {
			hostIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(hosts))))
			methodIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(methods))))
			accessLogs := sakurarslog2parquet.AccessLog{
				Host:       hosts[hostIndex.Int64()],
				RemoteAddr: "127.0.0.1",
				Identity:   "",
				User:       "",
				Time:       t,
				Method:     methods[methodIndex.Int64()],
				Path:       "/",
				Proto:      "HTTP/1.1",
				Status:     200,
				OutBytes:   1024 * 15,
				Referer:    "",
				UserAgent:  "test",
			}

			logJSON, err := json.Marshal(accessLogs)
			if err != nil {
				panic(err)
			}
			b.Write(logJSON)
			b.Write([]byte("\n"))
		}
		log.Printf("%s random logs: %.2fMiB", t.Format("2006/01/02 15:04"), float64(b.Len())/1024.0/1024.0)
		totalBytes += b.Len()

		key := fmt.Sprintf("%s%s.json", *prefix, t.Format("20060102T1504"))
		if err := upload(ctx, s3Client, key, &b); err != nil {
			panic(err)
		}
	}
	log.Printf("total: %.2fMiB", float64(totalBytes)/1024.0/1024.0)

	return nil
}

func loadConfig() (*Config, error) {
	f, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var conf Config
	if err := yaml.NewDecoder(f).Decode(&conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func newS3(conf *Config) *s3.Client {
	return s3.New(s3.Options{
		BaseEndpoint: &conf.S3.Endpoint,
		Region:       conf.S3.Region,
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     conf.S3.AccessKeyID,
				SecretAccessKey: conf.S3.SecretAccessKey,
			}, nil
		}),
	})
}

func upload(ctx context.Context, s3Client *s3.Client, key string, body io.Reader) error {
	s := time.Now()
	if _, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("log-test"),
		Key:    aws.String(key),
		Body:   body,
	}); err != nil {
		return err
	}
	elapsed := time.Since(s)
	log.Printf("%s update elapsed: %s", key, elapsed)
	return nil
}
