package bsmeter

import (
        "net/http"
        "log"
        html "code.google.com/p/go.net/html"
        "strings"
        "io/ioutil"
        "io"
        "regexp"
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

func downloadPage(url string) string {
    resp, err := http.Get(url)
    if err != nil {
        log.Printf("Got error : %v\n", err)
        return ""
    }
    defer resp.Body.Close()
    content, errRead := ioutil.ReadAll(resp.Body)
    if errRead != nil {
        log.Printf("Got error : %v\n", err)
        return ""
    }
    return string(content)
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
                        if isTitle {
                            title = cleanTitle(string(z.Text()))
                            continue
                        }
                        text := strings.ToLower(string(z.Text()))
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

