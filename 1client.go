package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	cfg  Config
	http *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{}}
}

func (c *Client) Fetch(path string, depth int) (*Page, error) {

	u := c.cfg.Addr + "/page?path=" + url.QueryEscape(path) + "&depth=" + strconv.Itoa(depth)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil { return nil, err }
	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }

	return decodePage(body)
}

func (c *Client) Push(p *Page) error {

	body, err := encodePage(p)
	if err != nil { return err }

	req, err := http.NewRequest(http.MethodPut, c.cfg.Addr+"/page", bytes.NewReader(body))
	if err != nil { return err }
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return fmt.Errorf("push: %s", resp.Status) }
	return nil
}

func (c *Client) Delete(path string) error {

	u := c.cfg.Addr + "/page?path=" + url.QueryEscape(path)
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil { return err }
	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return fmt.Errorf("delete: %s", resp.Status) }
	return nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	}
}
