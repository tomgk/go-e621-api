package e621api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

//This file implements the e621.net API part that's not available over the official API (or at least now known to be available there) and so is read from HTML

func getHtmlAttr(attrs []html.Attribute, name string) string {
	for _, attr := range attrs {
		if attr.Key == name {
			return attr.Val
		}
	}

	return ""
}

type BlacklistEntry struct {
	original []string
	positive []string
	negative []string
}

func (be BlacklistEntry) String() string {
	return strings.Join(be.original, " ")
}

func createBlacklistEntry(beTags []string) (e BlacklistEntry) {
	e.original = beTags

	for _, blacklisted := range beTags {
		if blacklisted[0] == '-' {
			e.negative = append(e.negative, blacklisted[1:])
		} else {
			e.positive = append(e.positive, blacklisted)
		}
	}

	return
}

func (be *BlacklistEntry) matches(tags []string) bool {
positive_search:
	for _, blacklisted := range be.positive {
		//all positive tags must be contained
		for _, tag := range tags {
			if tag == blacklisted {
				continue positive_search
			}
		}
		//not found => at least this one is missing
		return false
	}

	for _, blacklisted := range be.negative {
		for _, tag := range tags {
			//any negative tag means it's not blacklisted
			if tag == blacklisted {
				return false
			}
		}
	}

	//every positive tag is contained and no negative tag => it's blacklisted
	return true
}

func (api *E621Api) GetDefaultBlacklist() (blacklistedTags []BlacklistEntry, error *E621Error) {
	path := "posts"

	requestUrl := fmt.Sprintf(E621Url+"%s", path)

	log.Println("Requesting", requestUrl)

	rq, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		error = &E621Error{"Error creating request", err}
		return
	}
	rq.Header.Add("User-Agent", "tgbt (by vaddux on e621)")
	rq.Header.Add("Accept", "text/html")

	r, err := api.client.Do(rq)
	if err != nil {
		error = &E621Error{"Error requesting", err}
		return
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		error = &E621Error{"Response code", HttpStatus(r.StatusCode)}
		return
	}

	doc, err := html.Parse(r.Body)
	if err != nil {
		error = &E621Error{"Error parsing HTML", err}
		return
	}

	blacklistedTagsStr := ""

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" && getHtmlAttr(n.Attr, "name") == "blacklisted-tags" {
			blacklistedTagsStr = getHtmlAttr(n.Attr, "content")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if blacklistedTagsStr == "" {
		error = &E621Error{"No meta[name='blacklisted-tags']", nil}
		return
	}

	var tagLists []string
	if err := deserialize(strings.NewReader(blacklistedTagsStr), &tagLists); err != nil {
		fmt.Printf("ERROR: %+v\n", err)
		error = &E621Error{"Error parsing meta[name='blacklisted-tags'].content", err}
		return
	}

	blacklistedTags = make([]BlacklistEntry, len(tagLists))
	for i := range tagLists {
		blacklistedTags[i] = createBlacklistEntry(strings.Fields(tagLists[i]))
	}

	return
}
