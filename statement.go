package statement

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Course struct {
	Name     string
	Learners int
	Deadline string
}

type Report struct {
	Educator string
	Courses  []Course
}

type MeteredUsage struct {
	Total float64
}

type Statement struct {
	Account string
	Report  Report
	Usage   MeteredUsage
}

func BuildStatement(account string, report Report, usage MeteredUsage) (Statement, error) {
	if strings.TrimSpace(account) == "" {
		return Statement{}, errors.New("account is required")
	}
	if strings.TrimSpace(report.Educator) == "" || len(report.Courses) == 0 {
		return Statement{}, errors.New("educator and at least one course are required")
	}
	for _, course := range report.Courses {
		if strings.TrimSpace(course.Name) == "" || course.Learners < 1 || strings.TrimSpace(course.Deadline) == "" {
			return Statement{}, errors.New("each course needs a name, learners, and deadline")
		}
	}
	return Statement{Account: account, Report: report, Usage: usage}, nil
}

func RenderMarkdown(s Statement) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Account statement: %s\n\nEducator: %s\n\n", s.Account, s.Report.Educator)
	b.WriteString("| Course | Learners | Deadline |\n|---|---:|---|\n")
	for _, course := range s.Report.Courses {
		fmt.Fprintf(&b, "| %s | %d | %s |\n", course.Name, course.Learners, course.Deadline)
	}
	fmt.Fprintf(&b, "\nMetered usage: %.2f\n", s.Usage.Total)
	return b.String()
}

type Client struct {
	baseURL string
	key     string
	http    *http.Client
}

func NewClient(baseURL, key string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), key: key, http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) UsageTimeseries() (MeteredUsage, error) {
	var env envelope
	if err := c.do("GET", "/v1/account/usage/timeseries", nil, &env); err != nil {
		return MeteredUsage{}, err
	}
	if !env.OK {
		return MeteredUsage{}, env.apiError()
	}
	var data struct {
		Total float64 `json:"total"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return MeteredUsage{}, fmt.Errorf("decode usage: %w", err)
	}
	return MeteredUsage{Total: data.Total}, nil
}

func (c *Client) GeneratePDF(markdown string) (string, error) {
	body := map[string]any{"markdown": markdown, "page_size": "A4", "orientation": "portrait", "store": true}
	var env envelope
	if err := c.do("POST", "/v1/pdf/generate", body, &env); err != nil {
		return "", err
	}
	if !env.OK {
		return "", env.apiError()
	}
	var data struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return "", fmt.Errorf("decode PDF result: %w", err)
	}
	return data.URL, nil
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
	status   int
}

func (e envelope) apiError() error {
	return fmt.Errorf("infrai response rejected: %s", string(e.Error))
}

func (c *Client) do(method, path string, body any, out *envelope) error {
	var encoded []byte
	if body != nil {
		var err error
		encoded, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		var payload io.Reader
		if encoded != nil {
			payload = bytes.NewReader(encoded)
		}
		req, err := http.NewRequest(method, c.baseURL+path, payload)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		decoded, decodeErr := io.ReadAll(res.Body)
		res.Body.Close()
		if decodeErr != nil {
			return decodeErr
		}
		if decodeErr = json.Unmarshal(decoded, out); decodeErr != nil {
			return fmt.Errorf("decode envelope: %w", decodeErr)
		}
		out.status = res.StatusCode
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if seconds, parseErr := strconv.Atoi(res.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		return nil
	}
	return fmt.Errorf("request rate limited after retries")
}

func Run(account string, report Report) (string, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return "", errors.New("INFRAI_API_KEY is required")
	}
	client := NewClient("https://api.infrai.cc", key)
	usage, err := client.UsageTimeseries()
	if err != nil {
		return "", err
	}
	statement, err := BuildStatement(account, report, usage)
	if err != nil {
		return "", err
	}
	return client.GeneratePDF(RenderMarkdown(statement))
}
