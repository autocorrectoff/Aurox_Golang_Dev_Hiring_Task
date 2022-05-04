package engine

import (
	"testing"
)

// Full Flow Tests

func TestEngine_ShouldErrorWhenNoInputUrl(t *testing.T) {
	userInput := UserInput{
		Url:        "",
		Parallel:   5,
		OutputFile: "test_file.xml",
		MaxDepth:   5,
	}

	sm := New(userInput)
	err := sm.Start()
	if err == nil {
		t.Error("Expected Error got nil")
	}

}

func TestEngine_ShouldExtractLinksFromSitemapOrgWhenMaxDepthIs1(t *testing.T) {
	expected := []string{
		"https://www.sitemaps.org/faq.php",
		"https://www.sitemaps.org/protocol.php",
		"https://www.sitemaps.org/#",
		"https://www.sitemaps.org/terms.php",
		"http://creativecommons.org/licenses/by-sa/2.5/",
	}

	userInput := UserInput{
		Url:        "https://www.sitemaps.org", // Unlikely to change often which makes it good for tests
		Parallel:   1,
		OutputFile: "test_file.xml",
		MaxDepth:   1,
	}

	sm := New(userInput)
	err := sm.Start()
	if err != nil {
		t.Error("Expected Error got nil")
	}

	expectedLinksCount := 62

	if expectedLinksCount != len(sm.AccumulatedUrls) {
		t.Errorf("Expected Link count of %d but found %d", expectedLinksCount, len(sm.AccumulatedUrls))
	}

	for _, link := range expected {
		if !contains(sm.AccumulatedUrls, link) {
			t.Errorf("Link %s is missing", link)
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Unit Tests - TBD
