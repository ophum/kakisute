package simplemq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type SimpleMQClient interface {
	SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)
	ReceiveMessage(ctx context.Context) (*ReceiveMessageResponse, error)
	UpdateMessageTimeout(ctx context.Context, msgID string) error
	AckMessage(ctx context.Context, msgID string) error
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

type SendMessageResponse struct {
	Result  string   `json:"result"`
	Message *Message `json:"message"`
}

type ReceiveMessageResponse struct {
	Result   string     `json:"result"`
	Messages []*Message `json:"messages"`
}
type Message struct {
	ID                  string    `json:"id"`
	Content             string    `json:"content"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	ExpiresAt           time.Time `json:"expires_at"`
	AcquiredAt          time.Time `json:"acquired_at"`
	VisibilityTimeoutAt time.Time `json:"visibility_timeout_at"`
}

func (m *Message) UnmarshalJSON(b []byte) error {
	msg := struct {
		ID                  string `json:"id"`
		Content             string `json:"content"`
		CreatedAt           int64  `json:"created_at"`
		UpdatedAt           int64  `json:"updated_at"`
		ExpiresAt           int64  `json:"expires_at"`
		AcquiredAt          int64  `json:"acquired_at"`
		VisibilityTimeoutAt int64  `json:"visibility_timeout_at"`
	}{}
	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}

	m.ID = msg.ID
	m.Content = msg.Content
	m.CreatedAt = time.UnixMilli(msg.CreatedAt)
	m.UpdatedAt = time.UnixMilli(msg.UpdatedAt)
	m.ExpiresAt = time.UnixMilli(msg.ExpiresAt)
	m.AcquiredAt = time.UnixMilli(msg.AcquiredAt)
	m.VisibilityTimeoutAt = time.UnixMilli(msg.VisibilityTimeoutAt)
	return nil
}

type simpleMQClient struct {
	queueName  string
	token      string
	httpClient *http.Client
}

func NewSimpleMQClient(queueName, token string) SimpleMQClient {
	return &simpleMQClient{
		queueName:  queueName,
		token:      token,
		httpClient: http.DefaultClient,
	}
}

func (c *simpleMQClient) setHeader(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+c.token)
}

func (c *simpleMQClient) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(req); err != nil {
		return nil, err
	}
	log.Println("sendMessage", b.String())
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://simplemq.tk1b.api.sacloud.jp/v1/queues/%s/messages", c.queueName),
		&b,
	)
	if err != nil {
		return nil, err
	}

	c.setHeader(httpReq)

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpRes.Body)
		log.Println(string(body))
		return nil, errors.New(httpRes.Status)
	}
	log.Println(httpRes.Status)
	var res SendMessageResponse
	if err := json.NewDecoder(httpRes.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *simpleMQClient) ReceiveMessage(ctx context.Context) (*ReceiveMessageResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://simplemq.tk1b.api.sacloud.jp/v1/queues/%s/messages", c.queueName),
		nil,
	)
	if err != nil {
		return nil, err
	}
	c.setHeader(httpReq)

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		return nil, errors.New("fatal")
	}
	var res ReceiveMessageResponse
	if err := json.NewDecoder(httpRes.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

var ErrNotFound = errors.New("not found")

func (c *simpleMQClient) UpdateMessageTimeout(ctx context.Context, msgID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("https://simplemq.tk1b.api.sacloud.jp/v1/queues/%s/messages/%s", c.queueName, msgID),
		nil,
	)
	if err != nil {
		return err
	}
	c.setHeader(httpReq)

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		if httpRes.StatusCode == http.StatusNotFound {
			return ErrNotFound
		}
		return errors.New("fatal")
	}
	return nil
}

func (c *simpleMQClient) AckMessage(ctx context.Context, msgID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("https://simplemq.tk1b.api.sacloud.jp/v1/queues/%s/messages/%s", c.queueName, msgID),
		nil,
	)
	if err != nil {
		return err
	}
	c.setHeader(httpReq)

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		log.Println(httpRes)
		return errors.New("fatal")
	}
	return nil
}
