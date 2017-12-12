package crawler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// FUTURE TODO LIST
// ================
// 1. Incease test coverage
// 2. Include checking robots.txt
// 3. Include checking sitemap.xml when present
// 4. Implement trie data structure for quicker traversal and keeping track of navigation length between pages
// 5. Configure-drive the verbose level logging

var (
	// main data structure in use - hash map
	siteNodes = make(map[string]*SiteNode)

	rwMux = &sync.RWMutex{}
)

// NewGrawler returns a new Grawler instance
func NewGrawler(
	url *url.URL,
	maxConcurrency int,
	extractStaticAssets bool,
) *Grawler {
	return &Grawler{
		config: &grawlerConfig{
			url:                 url,
			maxConcurrency:      maxConcurrency,
			extractStaticAssets: extractStaticAssets,
		},
	}
}

// Grawler is the way to crawl urls
type Grawler struct {
	config *grawlerConfig
}

type grawlerConfig struct {
	maxConcurrency      int
	extractStaticAssets bool
	url                 *url.URL
}

// SiteNode is a struct that encapsulates the website node
type SiteNode struct {
	NodeURL              *url.URL
	IsCrawled            bool
	InternalStaticAssets map[string]bool // hash map will also serve as keeping the list unique. Ideally better to use a hash-set (not available in golang yet)
	InternalLinks        map[string]bool
}

func newSiteNode(url *url.URL) *SiteNode {
	return &SiteNode{
		NodeURL:              url,
		InternalStaticAssets: make(map[string]bool),
		InternalLinks:        make(map[string]bool),
	}
}

// Crawl crawls the url for links
func (g *Grawler) Crawl() map[string]*SiteNode {
	runtime.GOMAXPROCS(g.config.maxConcurrency)

	linksChan := make(chan *SiteNode)

	// Add landing page to list
	siteNodes[g.config.url.String()] = &SiteNode{
		NodeURL: g.config.url,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		crawl(g.config, g.config.url, linksChan, &wg)
	}()

	go func() {
		wg.Wait()
		close(linksChan)
	}()

	for l := range linksChan {
		select {
		default:
			if isSiteNodeDiscovered(l.NodeURL) {
				if l.IsCrawled {
					rwMux.Lock()
					siteNodes[l.NodeURL.String()].IsCrawled = l.IsCrawled
					siteNodes[l.NodeURL.String()].InternalStaticAssets = l.InternalStaticAssets
					siteNodes[l.NodeURL.String()].InternalLinks = l.InternalLinks
					rwMux.Unlock()

					log.Printf("Link crawling status updated to %v - %s", l.IsCrawled, l.NodeURL)
				}
			}
		}
	}

	return siteNodes
}

func getTokenizedHTML(u string) *html.Tokenizer {
	var tokenizedBody *html.Tokenizer

	resp, err := http.Get(u)
	if err != nil {
		log.Fatalln(err)
	} else if resp.StatusCode != http.StatusOK {
		return tokenizedBody
	}

	body := resp.Body
	// defer body.Close()

	tokenizedBody = html.NewTokenizer(body)

	return tokenizedBody
}

func crawl(config *grawlerConfig, urlToCrawl *url.URL, ch chan *SiteNode, wg *sync.WaitGroup) {
	log.Printf("Crawling started url:%s", urlToCrawl.String())

	defer wg.Done()

	siteNode := newSiteNode(urlToCrawl)

	resp, err := http.Get(urlToCrawl.String())
	if err != nil {
		log.Fatalln(err)
	} else if resp.StatusCode != http.StatusOK {
		return
	}

	body := resp.Body
	defer body.Close()

	// tokenizedBody := getTokenizedHTML(u.String())
	tokenizedBody := html.NewTokenizer(body)

	// TODO:: can parse the html faster with go routines ??
	for {
		// iterate over html tokens
		tokenType := tokenizedBody.Next()

		if tokenType == html.ErrorToken {
			// Encountered the end of html
			// log.Printf("Crawling ended url:%s\n", u.String())
			break
		} else if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			// handling cases for html nodes <a></a> and <a/>

			startToken := tokenizedBody.Token()

			// static asset can be img, script, css, link etc.
			newDiscoveredLink, isStaticAsset := extractLink(startToken, config.url.String())

			// if extracted link is empty, do nothing and move on
			if newDiscoveredLink == "" {
				continue
			}

			newDiscoveredLinkURL, err := url.Parse(newDiscoveredLink)
			if err != nil {
				// log.Fatalf("Error: %s - %s | %s", err.Error(), link, startToken.String())
				continue
			}

			// Check if the link found is internal i.e., under the target domain
			if !isLinkInternal(newDiscoveredLinkURL, getHostName(config.url)) {
				// log.Printf("Skipping as link [%s] is not under the target domain [%s]", l.Hostname(), getHostName(u))
				continue
			}

			if isStaticAsset && config.extractStaticAssets {
				// Track the internal static assets. No need to crawl.
				siteNode.InternalStaticAssets[newDiscoveredLink] = true
			} else if !isStaticAsset {
				// Normalize link i.e., strip away fragment & query
				newLinkNormalized := getNormalizedURL(newDiscoveredLinkURL)

				// Add to Internal links map
				siteNode.InternalLinks[newLinkNormalized.String()] = true

				// Check if the link is under target domain
				if !isSiteNodeDiscovered(newLinkNormalized) {
					log.Printf("New link discovered: %s [under %s]", newLinkNormalized.String(), urlToCrawl.String())

					addToSiteNodeList(newLinkNormalized.String(), newSiteNode(newLinkNormalized))

					// recursively crawl other web pages in the domain
					wg.Add(1)
					go func() {
						crawl(config, newLinkNormalized, ch, wg)
					}()
				}
			}
		}
	}

	siteNode.IsCrawled = true
	ch <- siteNode
}

// getNormalizedURL normalized the supplied url to be a page URL i.e., gets rid of any fragments
func getNormalizedURL(u *url.URL) *url.URL {
	u.RawQuery = ""
	u.Fragment = ""
	return u
}

func addToSiteNodeList(linkKey string, sn *SiteNode) {
	// In go 1.9, sync.Map should do the trick to handle concurrent map r/w
	rwMux.RLock()
	siteNodes[linkKey] = sn
	rwMux.RUnlock()
}

func isSiteNodeDiscovered(u *url.URL) bool {
	// In go 1.9, sync.Map should do the trick to handle concurrent map r/w
	rwMux.Lock()
	_, exists := siteNodes[u.String()]
	rwMux.Unlock()

	return exists
}

func extractLink(token html.Token, targetURL string) (string, bool) {
	switch strings.ToLower(token.Data) {
	case "a":
		return findAndGetHref(token, targetURL), false

	case "link":
		return findAndGetHref(token, targetURL), true

	case "script":
		return getSrc(token, targetURL), true

	case "img":
		return getSrc(token, targetURL), true
	}

	return "", false
}

func isLinkInternal(lu *url.URL, targetHostname string) bool {
	luh := getHostName(lu)
	return luh == targetHostname

	// luhParts := strings.Split(luh, ".")
	// luhLen := len(luhParts)

	// if luhLen == 2 {
	// 	return getHostName(lu) == targetHostname
	// } else if luhLen > 2 {
	// 	// handling sub-domain cases. For example: www.aws.amazon.com
	// 	// pick the last 2 segments that makes up our normalized hostname i.e., no scheme included
	// 	normalizedHn := strings.Join([]string{luhParts[luhLen-2], luhParts[luhLen-1]}, ".")
	// 	return normalizedHn == targetHostname
	// } else {
	// 	return false
	// }
}

func findAndGetHref(t html.Token, targetURL string) string {
	href := ""
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = extractAttribute(a, targetURL)
			break
		}
	}
	return href
}

func extractAttribute(a html.Attribute, baseURL string) string {
	href := ""

	if startsWith(a.Val, "http:") || startsWith(a.Val, "https:") {
		href = a.Val
	} else {
		u, err := url.Parse(a.Val)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			href = ""
		} else {
			base, err := url.Parse(baseURL)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				href = ""
			} else {
				href = base.ResolveReference(u).String()
			}
		}
	}

	// // for consistency sake
	// if len(href) > 0 && endsWith(href, "/") {
	// 	href += "/"
	// }

	return href
}

func getSrc(t html.Token, targetURL string) string {
	src := ""
	for _, a := range t.Attr {
		if a.Key == "src" {
			src = extractAttribute(a, targetURL)
			break
		}
	}
	return src
}

func getHostName(u *url.URL) string {
	h := u.Hostname()
	if ix := strings.Index(h, "www."); ix >= 0 {
		return h[4:]
	}
	return h
}

func startsWith(target, startsWith string) bool {
	return strings.Index(target, startsWith) == 0
}

func endsWith(target, endsWith string) bool {
	return strings.Index(target, endsWith) == len(target)-len(endsWith)
}
