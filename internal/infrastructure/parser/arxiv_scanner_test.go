package parser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"

	"ArticlesScanner/internal/scanner"
)

func TestBuildPageURL(t *testing.T) {
	t.Parallel()

	base := "https://export.arxiv.org/list/cs.AI/pastweek"
	u, err := buildPageURL(base, 200, 100)
	if err != nil {
		t.Fatalf("buildPageURL returned error: %v", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}

	if parsed.Scheme != "https" || parsed.Host != "export.arxiv.org" {
		t.Fatalf("unexpected host: %s", parsed.Host)
	}

	q := parsed.Query()
	if q.Get("skip") != "200" {
		t.Fatalf("expected skip=200, got %s", q.Get("skip"))
	}
	if q.Get("show") != "100" {
		t.Fatalf("expected show=100, got %s", q.Get("show"))
	}
}

func TestParseEntry(t *testing.T) {
	t.Parallel()

	html := `
	<dl>
	  <dt>
	    <span class="list-identifier"><a href="/abs/1234.56789">arXiv:1234.56789</a></span>
	  </dt>
	  <dd>
	    <div class="list-date">Date: 8 Nov 2025</div>
	    <div class="list-title mathjax">Title: Sample Title</div>
	    <p class="mathjax">Abstract: Sample abstract text.</p>
	  </dd>
	</dl>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("new document: %v", err)
	}

	dt := doc.Find("dt").First()
	dd := doc.Find("dd").First()

	article, publishedAt, err := parseEntry(dt, dd, "arxiv-ai", "cs.AI")
	if err != nil {
		t.Fatalf("parseEntry error: %v", err)
	}

	if article.ID != "arXiv:1234.56789" {
		t.Fatalf("unexpected id: %s", article.ID)
	}
	if article.Title != "Sample Title" {
		t.Fatalf("unexpected title: %s", article.Title)
	}
	if article.Abstract != "Sample abstract text." {
		t.Fatalf("unexpected abstract: %s", article.Abstract)
	}
	if article.Source != "arxiv-ai/cs.AI" {
		t.Fatalf("unexpected source: %s", article.Source)
	}

	wantDate := time.Date(2025, time.November, 8, 0, 0, 0, 0, time.UTC)
	if publishedAt.Format("2006-01-02") != wantDate.Format("2006-01-02") {
		t.Fatalf("unexpected published date: %v", publishedAt)
	}
}

func TestArxivScannerScan(t *testing.T) {
	t.Parallel()

	targetDay := time.Date(2025, time.November, 8, 0, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`
		<dl>
		  <dt>
		    <span class="list-identifier"><a href="/abs/2501.00001">arXiv:2501.00001</a></span>
		  </dt>
		  <dd>
		    <div class="list-date">Date: 8 Nov 2025</div>
		    <div class="list-title mathjax">Title: Fresh Article</div>
		    <p class="mathjax">Abstract: brand new.</p>
		  </dd>
		  <dt>
		    <span class="list-identifier"><a href="/abs/2501.00002">arXiv:2501.00002</a></span>
		  </dt>
		  <dd>
		    <div class="list-date">Date: 7 Nov 2025</div>
		    <div class="list-title mathjax">Title: Old Article</div>
		    <p class="mathjax">Abstract: older.</p>
		  </dd>
		</dl>`))
	}))
	defer server.Close()

	client := server.Client()
	sc := NewArxivScanner(client, nil)
	sc.pageSize = 10

	req := scanner.Request{
		Day:      targetDay,
		SiteName: "arxiv-ai",
		Categories: []scanner.Category{
			{Name: "cs.AI", URL: server.URL + "/list/cs.AI"},
		},
	}

	ctx := context.Background()
	articles, err := sc.Scan(ctx, req)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}

	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}

	if articles[0].ID != "arXiv:2501.00001" {
		t.Fatalf("unexpected article id: %s", articles[0].ID)
	}
	if articles[0].Abstract != "brand new." {
		t.Fatalf("unexpected abstract: %s", articles[0].Abstract)
	}
}
