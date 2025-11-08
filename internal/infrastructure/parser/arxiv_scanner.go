package parser

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"ArticlesScanner/internal/domain"
	"ArticlesScanner/internal/scanner"
)

const (
	arxivBaseURL = "https://arxiv.org"
)

var dateExpr = regexp.MustCompile(`\d{1,2} [A-Za-z]{3} \d{4}`)

// ArxivScanner crawls category pages and extracts articles for the requested day.
type ArxivScanner struct {
	client   *http.Client
	pageSize int
}

// NewArxivScanner wires an HTTP client; pageSize defaults to 200.
func NewArxivScanner(client *http.Client) *ArxivScanner {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &ArxivScanner{client: client, pageSize: 200}
}

// Name identifies the strategy inside the registry.
func (a *ArxivScanner) Name() string {
	return "arxiv"
}

// Scan walks through each category URL and returns all articles published on the requested day.
func (a *ArxivScanner) Scan(ctx context.Context, req scanner.Request) ([]domain.Article, error) {
	if len(req.Categories) == 0 {
		return nil, fmt.Errorf("no categories provided for site %s", req.SiteName)
	}

	targetDay := req.Day.UTC().Truncate(24 * time.Hour)
	results := make([]domain.Article, 0)
	seen := map[string]struct{}{}

	for _, cat := range req.Categories {
		skip := 0
		for {
			pageURL, err := buildPageURL(cat.URL, skip, a.pageSize)
			if err != nil {
				return nil, fmt.Errorf("category %s: %w", cat.Name, err)
			}

			doc, err := a.fetchDocument(ctx, pageURL)
			if err != nil {
				return nil, fmt.Errorf("category %s: %w", cat.Name, err)
			}

			pageArticles, shouldContinue := a.extractArticles(doc, targetDay, req.SiteName, cat.Name)
			for _, article := range pageArticles {
				if _, ok := seen[article.ID]; ok {
					continue
				}
				seen[article.ID] = struct{}{}
				results = append(results, article)
			}

			if !shouldContinue {
				break
			}
			skip += a.pageSize
		}
	}

	return results, nil
}

func (a *ArxivScanner) fetchDocument(ctx context.Context, pageURL string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "ArticlesScanner/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arxiv returned %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse document: %w", err)
	}

	return doc, nil
}

func (a *ArxivScanner) extractArticles(doc *goquery.Document, targetDay time.Time, siteName, category string) ([]domain.Article, bool) {
	var (
		collected    []domain.Article
		continueScan = true
		processed    int
	)

	doc.Find("dl > dt").EachWithBreak(func(i int, dt *goquery.Selection) bool {
		dd := dt.Next()
		processed++

		article, publishedAt, err := parseEntry(dt, dd, siteName, category)
		if err != nil {
			return true
		}

		articleDay := publishedAt.UTC().Truncate(24 * time.Hour)
		if articleDay.Equal(targetDay) {
			collected = append(collected, article)
		}
		if articleDay.Before(targetDay) {
			continueScan = false
			return false
		}

		return true
	})

	if processed < a.pageSize {
		continueScan = false
	}

	return collected, continueScan
}

func parseEntry(dt, dd *goquery.Selection, siteName, category string) (domain.Article, time.Time, error) {
	var article domain.Article

	id := strings.TrimSpace(dt.Find("a[href*=\"/abs/\"]").First().Text())
	if id == "" {
		if href, exists := dt.Find("a[href*=\"/abs/\"]").First().Attr("href"); exists {
			id = strings.TrimPrefix(href, "/abs/")
		}
	}

	link := dt.Find("a[href*=\"/abs/\"]").First()
	href, _ := link.Attr("href")
	if !strings.HasPrefix(href, "http") {
		href = strings.TrimSuffix(arxivBaseURL, "/") + href
	}

	title := strings.TrimSpace(dd.Find(".list-title").First().Text())
	title = strings.TrimPrefix(title, "Title:")
	title = strings.TrimSpace(title)

	summary := dd.Find(".mathjax").First().Text()
	summary = strings.TrimPrefix(summary, "Abstract:")
	summary = strings.TrimSpace(summary)

	dateText := strings.TrimSpace(dd.Find(".list-date").First().Text())
	if dateText == "" {
		dateText = strings.TrimSpace(dd.Find(".list-dateline").First().Text())
	}

	match := dateExpr.FindString(dateText)
	publishedAt := time.Now().UTC()
	if match != "" {
		if parsed, err := time.Parse("2 Jan 2006", match); err == nil {
			publishedAt = parsed
		}
	}

	if id == "" {
		id = href
	}

	source := siteName
	if category != "" {
		source = fmt.Sprintf("%s/%s", siteName, category)
	}

	article = domain.Article{
		ID:          id,
		Title:       title,
		Abstract:    summary,
		URL:         href,
		Source:      source,
		PublishedAt: publishedAt,
	}

	return article, publishedAt, nil
}

func buildPageURL(base string, skip, pageSize int) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid category url %s: %w", base, err)
	}

	query := parsed.Query()
	query.Set("skip", strconv.Itoa(skip))
	query.Set("show", strconv.Itoa(pageSize))
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
