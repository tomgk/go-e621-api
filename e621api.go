package e621api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type E621Error struct {
	Type         string
	WrappedError error
}

func (e *E621Error) Error() string {
	if e.WrappedError != nil {
		return fmt.Sprintf("%v: %s", e.Type, e.WrappedError)
	} else {
		return e.Type
	}
}

func (e *E621Error) Unwrap() error {
	return e.WrappedError
}

var _ error = &E621Error{}

//turns "this_text" into "this text"
//this escaping is used on tags for example
func TextifyEscaped(escaped string) string {
	return strings.ReplaceAll(escaped, "_", " ")
}

type HttpStatus int

func (h HttpStatus) Error() string {
	return fmt.Sprintf("HTTP status %v", int(h))
}

type E621Api struct {
	client *http.Client
}

func CreateE621Api() *E621Api {
	return &E621Api{client: &http.Client{}}
}

const E621Url = "https://e621.net/"

func (api *E621Api) Request(path string, result interface{}) (error *E621Error) {
	requestUrl := fmt.Sprintf(E621Url+"%s", path)

	log.Println("Requesting", requestUrl)

	rq, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		error = &E621Error{"Error creating request", err}
		return
	}
	rq.Header.Add("User-Agent", "tgbt (by vaddux on e621)")
	rq.Header.Add("Accept", "application/json")

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

	debug := true
	if debug {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			error = &E621Error{"Error reading", err}
			return
		}
		//fmt.Printf("Response for %v: %s\n", requestUrl, data)

		dumpFileName := url.QueryEscape(requestUrl)
		maxLen := 64
		if len(dumpFileName) > maxLen {
			dumpFileName = dumpFileName[len(dumpFileName)-maxLen:]
		}

		dumpFileName = "cache/" + dumpFileName + ".json"

		ioutil.WriteFile(dumpFileName, data, 0644)
		log.Printf("Response for %v written to %s\n", requestUrl, dumpFileName)

		r.Body = ioutil.NopCloser(bytes.NewReader(data))
	}
	src := r.Body

	error = deserialize(src, result)

	return
}

func deserialize(src io.Reader, result interface{}) (error *E621Error) {
	dec := json.NewDecoder(src)
	//dec.DisallowUnknownFields()

	if err := dec.Decode(result); err != nil {
		error = &E621Error{"Error parsing", err}
		return
	}

	return
}

type MetaTagged interface {
	GetMetaTag(name string) (val string, ok bool)
}

type Post struct {
	ID      int                 `json:"id"`
	Tags    map[string][]string `json:"tags"`
	File    PostFile            `json:"file"`
	Sample  SampleFile          `json:"sample"`
	Preview PreviewFile         `json:"preview"`
	Rating  string              `json:"rating"`
}

func (post *Post) Url() string {
	return fmt.Sprintf("https://e621.net/posts/%v", post.ID)
}

func (post *Post) GetMetaTag(name string) (val string, ok bool) {
	ok = true

	if name == "id" {
		val = fmt.Sprintf("%v", post.ID)
	} else if name == "md5" {
		val = post.File.MD5
	} else {
		ok = false
	}

	//TODO: many are missing

	return
}

type PostFile struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Ext    string `json:"ext"`
	Size   int64  `json:"size"`
	MD5    string `json:"md5"`
}

type SampleFile struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type PreviewFile struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Constructs an file URL from the data if url is not set for some reason
func (pf *PostFile) GetUrl() string {
	if pf.Url != "" {
		return pf.Url
	}

	log.Printf("WARN: post.file.url not set\n")

	if pf.MD5 == "" {
		return ""
	}

	url := "https://static1.e621.net/data/"

	md5 := pf.MD5

	url = url + md5[0:2] + "/" + md5[2:4] + "/" + md5 + "." + pf.Ext
	return url
}

type PostPage struct {
	Post Post `json:"post"`
}

func (api *E621Api) GetPost(id int) (post Post, error *E621Error) {
	var p PostPage
	url := fmt.Sprintf("posts/%v", id)
	if err := api.Request(url, &p); err != nil {
		error = err
		return
	}
	post = p.Post
	return
}

func (api *E621Api) GetRandomPost() (post Post, error *E621Error) {
	var p PostPage
	url := fmt.Sprintf("posts/random")
	if err := api.Request(url, &p); err != nil {
		error = err
		return
	}
	post = p.Post
	return
}

type PostsPage struct {
	Posts []Post `json:"posts"`
}

type PostSearch struct {
	tags  []string
	page  int
	limit int
}

func (api *E621Api) GetPosts(search PostSearch) (posts []Post, error *E621Error) {
	var p PostsPage

	url := fmt.Sprintf("posts?tags=%s", url.QueryEscape(strings.Join(search.tags, " ")))

	if search.page != 0 {
		url += fmt.Sprintf("&page=%v", search.page)
	}

	if search.limit != 0 {
		url += fmt.Sprintf("&limit=%v", search.limit)
	}

	if err := api.Request(url, &p); err != nil {
		error = err
		return
	}
	posts = p.Posts
	return
}

type Pool struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PostIds     []int  `json:"post_ids"`
	PostCount   int    `json:"post_count"`
}

func (p *Pool) GetUrl() string {
	return fmt.Sprintf(E621Url+"pools/%v", p.ID)
}

func (api *E621Api) GetPool(poolId int) (pool Pool, error *E621Error) {
	url := fmt.Sprintf("pools/%v.json", poolId)

	if err := api.Request(url, &pool); err != nil {
		error = err
		return
	}

	return
}

func (api *E621Api) GetPoolPosts(pool Pool) (posts []Post, error *E621Error) {
	tags := []string{fmt.Sprintf("pool:%v", pool.ID)}

	posts, error = api.GetPosts(PostSearch{tags: tags, limit: pool.PostCount})
	if error != nil {
		return
	}

	sortPosts(pool.PostIds, posts)
	return
}

func indexOf(nums []int, n int) int {
	for idx, x := range nums {
		if x == n {
			return idx
		}
	}

	return len(nums) //put at end
}

func sortPosts(orderedPostIds []int, posts []Post) {
	sort.Slice(posts, func(a, b int) bool {
		return indexOf(orderedPostIds, posts[a].ID) < indexOf(orderedPostIds, posts[b].ID)
	})
}
