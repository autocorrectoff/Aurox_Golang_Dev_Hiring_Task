package engine

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/autocorrectoff/SimpleSitemapGenerator/utils"
)

type SiteMap struct {
	Url             string
	Parallel        int
	OutputFile      string
	MaxDepth        int
	BaseUrl         string
	VisitedUrls     map[string]int
	AccumulatedUrls []string
	mu              sync.RWMutex
}

type UserInput struct {
	Url        string
	Parallel   int
	OutputFile string
	MaxDepth   int
}

type HTTPResponse struct {
	url      string
	response *http.Response
	body     string
	err      error
}

func New(config UserInput) *SiteMap {
	return &SiteMap{
		Url:             config.Url,
		Parallel:        config.Parallel,
		OutputFile:      config.OutputFile,
		MaxDepth:        config.MaxDepth,
		VisitedUrls:     make(map[string]int),
		AccumulatedUrls: []string{},
		mu:              sync.RWMutex{},
	}
}

func (sm *SiteMap) Start() error {
	if sm.Url == "" {
		return errors.New("we can't do this without an URL")
	}
	sm.BaseUrl = sm.getBaseUrl()
	httpResponse := sm.fetchPageGuard(sm.Url)
	urlFromBaseTag := checkForBaseTag(httpResponse.body)
	if urlFromBaseTag != "" {
		sm.BaseUrl = urlFromBaseTag
	}
	links := extractLinksFromHtml(httpResponse.body)
	links = sm.prependBaseUrlIfMissing(links)
	sm.appendFetchedLinks(links)
	sm.recurse(links, 0)
	sm.AccumulatedUrls = utils.RemoveDuplicateStr(sm.AccumulatedUrls)
	return nil
}

func (sm *SiteMap) Export() {
	xmlTag := `<?xml version="1.0" encoding="UTF-8"?>`
	urlSetOpeningTag := `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"	xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">`
	urlSetClosingTag := `</urlset>`
	urlOpeningTag := `<url>`
	urlClosingTag := `</url>`
	locOpeningTag := `<loc>`
	locClosingTag := `</loc>`
	var urlTags string
	for _, link := range sm.AccumulatedUrls {
		urlTags += (urlOpeningTag + locOpeningTag + link + locClosingTag + urlClosingTag)
	}
	xmlContent := xmlTag + urlSetOpeningTag + urlTags + urlSetClosingTag
	file, err := os.Create(sm.OutputFile)
	utils.HandleError(err)
	file.WriteString(xmlContent)
	file.Close()
}

func (sm *SiteMap) getBaseUrl() string {
	parts := strings.SplitAfter(sm.Url, "/")
	baseUrl := parts[0] + parts[1] + parts[2]
	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl = baseUrl + "/"
	}
	return baseUrl
}

func (sm *SiteMap) recurse(links []string, currentIteration int) {
	if currentIteration > sm.MaxDepth {
		return
	}
	if len(links) == 0 {
		return
	}
	childLinks := sm.fetchSubsequent(links)
	sm.appendFetchedLinks(childLinks)
	currentIteration++
	sm.recurse(childLinks, currentIteration)
}

func (sm *SiteMap) fetchSubsequent(links []string) []string {
	urlSlices := utils.SplitToChunks(links, sm.Parallel)
	ch := make(chan []string, len(urlSlices))
	var extracted [][]string
	for _, linkSlice := range urlSlices {
		go func(linkSlice []string) {
			var parsedLinks []string
			for _, link := range links {
				visited := sm.handleVisitedUrls(link)
				if !visited {
					httpResp := sm.fetchPageGuard(link)
					if httpResp != nil {
						childLinks := extractLinksFromHtml(httpResp.body)
						childLinks = sm.prependBaseUrlIfMissing(childLinks)
						parsedLinks = append(parsedLinks, childLinks...)
					}
				}
			}
			ch <- parsedLinks
		}(linkSlice)
	}

	for {
		select {
		case r := <-ch:
			extracted = append(extracted, r)
			if len(extracted) == len(urlSlices) {
				linksSlice, err := utils.FlattenDepthString(reflect.ValueOf(extracted), 2)
				utils.HandleError(err)
				return linksSlice
			}
		case <-time.After(50 * time.Millisecond):
			fmt.Printf(".")
		}
	}

}

func (sm *SiteMap) appendFetchedLinks(links []string) {
	sm.AccumulatedUrls = append(sm.AccumulatedUrls, links...)
}

func (sm *SiteMap) prependBaseUrlIfMissing(links []string) []string {
	var fullLinks []string
	for _, link := range links {
		var fullLink string
		if !strings.HasPrefix(link, "http") {
			fullLink = sm.BaseUrl + link
			fullLinks = append(fullLinks, fullLink)
		} else {
			fullLinks = append(fullLinks, link)
		}
	}
	return fullLinks
}

// Some links are external web pages
func (sm *SiteMap) fetchPageGuard(url string) *HTTPResponse {
	urlCopy := url
	if !strings.HasSuffix(url, "/") {
		urlCopy = url + "/"
	}
	if strings.HasPrefix(urlCopy, sm.BaseUrl) {
		resp := fetchPage(url)
		if resp != nil && resp.err != nil {
			return nil
		}
		return resp
	}
	return nil
}

func (sm *SiteMap) handleVisitedUrls(url string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	visitCount := sm.VisitedUrls[url]
	if visitCount > 0 {
		return true
	}
	visitCount++
	sm.VisitedUrls[url] = visitCount
	return false
}

func fetchPage(url string) *HTTPResponse {
	fmt.Printf("Fetching %s \n", url)
	resp, err := http.Get(url)
	utils.HandleError(err)
	if resp != nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		utils.HandleError(err)
		bs := string(body)
		return &HTTPResponse{url, resp, bs, err}
	}
	return nil
}

func checkForBaseTag(html string) string {
	r, _ := regexp.Compile(`<base.+?\s*href\s*=\s*["\']?([^"\'\s>]+)["\']?`)
	baseTag := r.FindString(html)
	if baseTag != "" {
		clean := strings.Replace(baseTag, "<base href=", "", -1)
		clean = strings.Replace(clean, `"`, "", -1)
		clean = strings.Replace(clean, `'`, "", -1)
		return clean
	}
	return ""
}

// Since we're using only standard libraries we're using regexp here to find anchor tags
// Using some html DOM library would be a much better option
// Like: github.com/PuerkitoBio/goquery
func extractLinksFromHtml(html string) []string {
	r, _ := regexp.Compile(`<a.+?\s*href\s*=\s*["\']?([^"\'\s>]+)["\']?`)
	result := r.FindAllString(html, -1)
	var links []string
	for _, link := range result {
		clean := strings.Replace(link, "<a href=", "", -1)
		clean = strings.Replace(clean, `"`, "", -1)
		clean = strings.Replace(clean, `'`, "", -1)
		clean = removeQueryString(clean)
		if strings.HasPrefix(clean, "http") || !strings.HasPrefix(clean, "<a") {
			links = append(links, clean)
		}
	}
	return links
}

func removeQueryString(url string) string {
	return strings.Split(url, "?")[0]
}
