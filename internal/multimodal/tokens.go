package multimodal

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/ikhsan3adi/gemini-web2api/internal/config"
	"github.com/ikhsan3adi/gemini-web2api/internal/gemini"
)

type PageTokens struct {
	PushID string
	Pctx   string
	At     string
}

type TokenCache struct {
	mu     sync.Mutex
	cfg    config.Config
	cookie *gemini.CookieCache
	client *http.Client
	ts     time.Time
	tokens PageTokens
}

func NewTokenCache(cfg config.Config, cookie *gemini.CookieCache, client *http.Client) *TokenCache {
	return &TokenCache{
		cfg:    cfg,
		cookie: cookie,
		client: client,
		tokens: PageTokens{
			PushID: "feeds/mcudyrk2a4khkz",
			Pctx:   "CgcSBWjK7pYx",
			At:     "",
		},
	}
}

func (c *TokenCache) fetchPageTokens() PageTokens {
	tokens := PageTokens{
		PushID: "feeds/mcudyrk2a4khkz",
		Pctx:   "CgcSBWjK7pYx",
		At:     "",
	}

	reqURL := fmt.Sprintf("https://gemini.google.com%s/app", gemini.AccountPrefix(c.cfg.AuthUser))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return tokens
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	cookieInfo, _ := c.cookie.Load()
	if cookieInfo.Cookie != "" {
		req.Header.Set("Cookie", cookieInfo.Cookie)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Page token fetch failed: %v", err)
		return tokens
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokens
	}

	html := string(bodyBytes)
	if m := regexp.MustCompile(`"qKIAYe":"([^"]+)"`).FindStringSubmatch(html); len(m) > 1 {
		tokens.PushID = m[1]
	}
	if m := regexp.MustCompile(`"Ylro7b":"([^"]+)"`).FindStringSubmatch(html); len(m) > 1 {
		tokens.Pctx = m[1]
	}
	if m := regexp.MustCompile(`"thykhd":"([^"]+)"`).FindStringSubmatch(html); len(m) > 1 {
		tokens.At = m[1]
	}

	return tokens
}

func (c *TokenCache) Get() PageTokens {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.ts) > 600*time.Second {
		c.tokens = c.fetchPageTokens()
		c.ts = time.Now()
	}

	return c.tokens
}
