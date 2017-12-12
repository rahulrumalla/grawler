package crawler

import (
	"fmt"
	"log"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartsWith(t *testing.T) {
	res := startsWith("www.hello.com", "www")
	assert.True(t, res, "Expected true")

	res = startsWith("www.hello.com", "http")
	assert.False(t, res, "Expected false")
}

func TestEndsWith(t *testing.T) {
	res := endsWith("www.hello.com", "com")
	assert.True(t, res, "Expected true")

	res = endsWith("www.hello.com", "http")
	assert.False(t, res, "Expected false")
}

func TestGetHostname(t *testing.T) {
	testCases := make(map[string]string)
	testCases["https://www.rahulrumalla.com"] = "rahulrumalla.com"
	testCases["https://rahulrumalla.com"] = "rahulrumalla.com"
	testCases["http://localhost:5454"] = "localhost"
	testCases["http://localhost:5454/style.css"] = "localhost"

	for input, expected := range testCases {
		u, err := url.Parse(input)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			t.Fail()
		}

		actual := getHostName(u)
		assert.Equal(t, expected, actual, "Expected and actual hostname are not the same")
	}
}

func TestGetNormalizedUrl(t *testing.T) {
	testCases := make(map[string]string)
	testCases["https://monzo.com/cdn-cgi/l/email-protection#355d50594575585a5b4f5a1b565a58"] = "https://monzo.com/cdn-cgi/l/email-protection"
	testCases["https://rahulrumalla.com?id=123#section"] = "https://rahulrumalla.com"

	for input, expected := range testCases {
		u, err := url.Parse(input)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			t.Fail()
		}

		result := getNormalizedURL(u)
		actual := result.String()

		assert.Equal(t, expected, actual, "Excted and actual are not the same")
	}
}

func TestCanGetTokenizedHTML(t *testing.T) {
	tb := getTokenizedHTML("http://localhost:5454")
	assert.NotNil(t, tb, "Nil body")
	assert.Nil(t, tb.Err(), "Error")
}

func TestCanCrawlExcludingStaticAssets(t *testing.T) {
	u, err := url.ParseRequestURI("http://localhost:5454")
	if err != nil {
		log.Printf("Error: %s", err.Error())
		log.Printf("Make sure to run the test web server found at server.go") // Should automate this is testSetup() and teardown()
		t.Fail()
	}

	expectedNumberOfLinks := 4

	g := NewGrawler(u, 4, false)
	siteNodes := g.Crawl()

	for k, v := range siteNodes {
		fmt.Println(k)

		assert.True(t, v.IsCrawled, "The site node is not crawled!")

		for a := range v.InternalLinks {
			fmt.Printf("  └── %s\n", a)
		}
	}

	assert.NotEmpty(t, siteNodes, "Empty site nodes")
	assert.Equal(t, expectedNumberOfLinks, len(siteNodes), "Expected and actual number of site nodes are different")
}

func TestCanCrawlIncludingStaticAssets(t *testing.T) {
	u, err := url.ParseRequestURI("http://localhost:5454")
	if err != nil {
		log.Printf("Error: %s", err.Error())
		log.Printf("Make sure to run the test web server found at ./server.go") // Should automate this is testSetup() and teardown()
		t.Fail()
	}

	expectedNumberOfLinks := 7

	g := NewGrawler(u, 4, true)
	siteNodes := g.Crawl()

	counter := 0
	for k, v := range siteNodes {
		fmt.Println(k)
		counter++

		assert.True(t, v.IsCrawled, "The site node is not crawled!")

		for a := range v.InternalStaticAssets {
			counter++
			fmt.Printf("  └── %s\n", a)
		}
	}

	assert.NotEmpty(t, siteNodes, "Empty site nodes")
	assert.Equal(t, expectedNumberOfLinks, counter, "Expected and actual number of site nodes are different")
}
