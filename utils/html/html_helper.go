package html

import (
	html "code.google.com/p/go.net/html"
	publicsuffix "code.google.com/p/go.net/publicsuffix"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	bstrings "github.com/bonnefoa/gobot/utils/strings"
)

var reUrls, _ = regexp.Compile("https?://[^ ]*")

func ExtractUrls(mess string) []string {
	urls := reUrls.FindAllString(mess, -1)
	for i := range urls {
		if strings.Contains(urls[i], "imgur") {
			urls[i] = strings.TrimSuffix(urls[i], ".jpeg")
			urls[i] = strings.TrimSuffix(urls[i], ".jpg")
		}
	}
	return urls
}

func cleanTitle(title string) string {
	title = strings.Replace(title, "\n", "", -1)
	title = strings.TrimSpace(title)
	return title
}

func DownloadPage(url string) ([]byte, error) {
	opts := &cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(opts)
	if err != nil {
		log.Printf("Error on jar creation : %v\n", err)
		return []byte{}, err
	}
	client := &http.Client{Jar: jar}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Got error : %v\n", err)
		return []byte{}, err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Got error : %v\n", err)
		return []byte{}, err
	}
	return content, nil
}

func TokenizePage(r io.Reader) ([]string, string) {
	res := []string{}
	z := html.NewTokenizer(r)
	isTitle := false
	title := ""
loop:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break loop
		case html.TextToken:
			text := string(z.Text())
			if isTitle {
				title = cleanTitle(text)
				continue
			}
			res = append(res, bstrings.TokenizeWords(text)...)
		case html.EndTagToken:
			tn, _ := z.TagName()
			if string(tn) == "title" {
				isTitle = false
			}
		case html.StartTagToken:
			tn, _ := z.TagName()
			if string(tn) == "title" {
				isTitle = true
			}
		}
	}
	return res, title
}
