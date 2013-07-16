package bsmeter

import (
        "net/http"
        "log"
        "regexp"
        html "code.google.com/p/go.net/html"
        "strings"
)

var reUrls, _ = regexp.Compile("https?://[^ ]*")

func ExtractUrls(mess string) []string {
        urls := reUrls.FindAllString(mess, -1)
        for i := range urls {
            urls[i] = strings.TrimSuffix(urls[i], ".jpeg")
            urls[i] = strings.TrimSuffix(urls[i], ".jpg")
        }
        return urls
}

func cleanTitle(title string) string {
    title = strings.Replace(title, "\n", "", -1)
    title = strings.TrimSpace(title)
    return title
}

func LookupTitle(url string) string {
        resp, err := http.Get(url)
        if err != nil {
                log.Printf("Got error is %v\n", err)
                return ""
        }
        z := html.NewTokenizer(resp.Body)
        isTitle := false
        res := ""

        for {
                tt := z.Next()
                switch tt {
                case html.ErrorToken:
                        return ""
                case html.TextToken:
                        if isTitle {
                                res = string(z.Text())
                        }
                case html.StartTagToken:
                        tn, _ := z.TagName()
                        if string(tn) == "title" {
                                isTitle = true
                        }
                case html.EndTagToken:
                        tn, _ := z.TagName()
                        if string(tn) == "title" {
                                resp.Body.Close()
                                return cleanTitle(res)
                        }
                }
        }
        return ""
}
