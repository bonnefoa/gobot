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
)

var reUrls, _ = regexp.Compile("https?://[^ ]*")
var notWords, _ = regexp.Compile("[^a-zA-Z]+")

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

func DownloadPage(url string) []byte {
	opts := &cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(opts)
	if err != nil {
		log.Printf("Error on jar creation : %v\n", err)
		return []byte{}
	}
	client := &http.Client{Jar: jar}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Got error : %v\n", err)
		return []byte{}
	}
	defer resp.Body.Close()
	content, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		log.Printf("Got error : %v\n", err)
		return []byte{}
	}
	return content
}

func TokenizeWords(text string) []string {
	res := []string{}
	text = strings.ToLower(text)
	text = strings.Replace(text, "\n", "", -1)
	for _, word := range strings.Split(text, " ") {
		word = strings.TrimSpace(word)
		if notWords.FindAllString(word, 1) != nil {
			continue
		}
		if word != "" {
			res = append(res, word)
		}
	}
	return res
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
			res = append(res, TokenizeWords(text)...)
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
