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

func LookupTitle(url string) (string, bool) {
        log.Printf("Lookup title for url %s\n", url)
        resp, err := http.Get(url)
        if err != nil {
                log.Printf("Got error : %v\n", err)
                return "", false
        }
        defer resp.Body.Close()
        z := html.NewTokenizer(resp.Body)
        isTitle := false

        for {
                tt := z.Next()
                switch tt {
                case html.ErrorToken:
                        return "", false
                case html.TextToken:
                        if isTitle {
                                title := cleanTitle(string(z.Text()))
                                if title != "" {
                                    return title, true
                                }
                        }
                case html.StartTagToken:
                        tn, _ := z.TagName()
                        if string(tn) == "title" {
                                isTitle = true
                        }
                }
        }
        return "", false
}
