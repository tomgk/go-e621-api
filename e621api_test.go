package e621api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	//"net/http"
	//"fmt"
	//"io"
	//"io/ioutil"
	//"strings"
)

func Test_sortPosts(t *testing.T) {
	indexes := []int{50, 40, 30, 20, 10}
	posts := []Post{
		Post{ID: 10},
		Post{ID: 20},
		Post{ID: 30},
		Post{ID: 50},
		Post{ID: 40},
		Post{ID: 15},
	}

	sortPosts(indexes, posts)

	result := make([]int, len(posts))
	for idx, post := range posts {
		result[idx] = post.ID
	}

	indexes = append(indexes, 15) //the one not contained int indexes

	assert.Equal(t, indexes, result, "")
}

func Test_PostFile_GetUrl(t *testing.T) {
	fileWithUrl := &PostFile{Url: "https://static1.e621.net/data/e2/4e/e24e3ce9944afc88e8c8204dd279940d.png"}
	fileWithoutUrl := &PostFile{MD5: "e24e3ce9944afc88e8c8204dd279940d", Ext: "png"}

	assert.Equal(t, "https://static1.e621.net/data/e2/4e/e24e3ce9944afc88e8c8204dd279940d.png", fileWithUrl.GetUrl(), "")

	assert.Equal(t, "https://static1.e621.net/data/e2/4e/e24e3ce9944afc88e8c8204dd279940d.png", fileWithoutUrl.GetUrl(), "")
}

func Test_TextifyEscaped(t *testing.T) {
	assert.Equal(t, "this text", TextifyEscaped("this_text"), "")
	assert.Equal(t, "text", TextifyEscaped("text"), "")
	assert.Equal(t, "", TextifyEscaped(""), "")
}

func TestPoolGetUrl(t *testing.T) {
	p := &Pool{ID: 1234}
	assert.Equal(t, "https://e621.net/pools/1234", p.GetUrl(), "")
}

/*
type Faked struct{
  data map[string]string
}

func makeFaked() *Faked{
  return &Faked{make(map[string]string)}
}

func (f *Faked) RoundTrip(req *http.Request) (*http.Response, error) {
  resp := f.data[req.URL.String()]

  status := "200 OK"
  statusCode := 200
  var contentLength int64 = 0

  var body io.ReadCloser

  if resp != "" {
    body = ioutil.NopCloser(strings.NewReader(resp))
    contentLength = int64(len(resp))
  } else {
    status = "404 Not Found"
    statusCode = 404
  }

  return &http.Response{
    Status: status,
    StatusCode: statusCode,
    Proto: "HTTP/1.0",
    ProtoMajor: 1,
    ProtoMinor: 0,
    Header: http.Header{},
    Body: body,
    ContentLength: contentLength,
    TransferEncoding: nil,
    Close: false,
    Uncompressed: false,
    Trailer: nil,
    Request: req,
    TLS: nil,
  }, nil
}

func Test_PoolDecode(t *testing.T) {
  //Pool 1234
  //initially just used because it's 1234
  //turns out that most of the posts couldn't be read for some reason, that's why it's used for a test now
  //UPDATE: Turns out some post don't have post.file.url set for some reason

  faked := makeFaked()

  faked.data["https://e621.net/pools/1234.json"] = ...
  faked.data["https://e621.net/posts?tags=pool%3A1234&limit=7"] = ...

  hc := &http.Client{}
  hc.Transport = faked

  v := E621Api{hc}

  pool, err := v.GetPool(1234)
  AssertEqErr(t, nil, err, "")

  posts, err := v.GetPoolPosts(pool)
  AssertEqErr(t, nil, err, "")

  fmt.Printf("%+v", posts)

  //expectedPosts = []Post{
  //  ...
  //}
}

func AssertEqErr(t *testing.T, expected, actual *E621Error, msg string){
  if msg == "" {
    if expected == nil {
      msg = fmt.Sprintf("Actual error as string: %v", actual)
    }
  }

  assert.Equal(t, expected, actual, msg)
}
*/
