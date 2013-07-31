package meteo

import (
        uhtml "utils/html"
        html "code.google.com/p/go.net/html"
        "io"
        "strings"
        "bytes"
        "fmt"
        "regexp"
        "time"
        "irc/message"
        "log"
)

var reRain, _ = regexp.Compile("Pluie [fm]")

func hasRain(parsed []string) bool {
        for _, el := range parsed {
                if reRain.MatchString(el) {
                        return true
                }
        }
        return false
}

func hasTablPluieClass(z *html.Tokenizer) bool{
        key, val, more := z.TagAttr()
        if string(key) == "class" && string(val) == "tablPluie" {
                return true
        }
        if more{
                return hasTablPluieClass(z)
        }
        return false
}

func ParseWeather(r io.Reader) []string{
        res := []string{}
        z := html.NewTokenizer(r)
        inTablePluie := false
        candidateText := false
        horaire := ""
        loop:
        for {
                tt := z.Next()
                switch tt {
                case html.ErrorToken:
                        break loop
                case html.TextToken:
                        if candidateText {
                                text := strings.TrimSpace(string(z.Text()))
                                if text != "" {
                                        if horaire == "" {
                                                horaire = text
                                        } else {
                                                res = append(res,
                                                        fmt.Sprintf("%s : %s", horaire, text))
                                                horaire = ""
                                        }
                                }
                        }
                case html.EndTagToken:
                        candidateText = false
                        if hasTablPluieClass(z) {
                                return res
                        }
                        break
                case html.StartTagToken:
                        if !inTablePluie && hasTablPluieClass(z) {
                                inTablePluie = true
                        } else if inTablePluie {
                                tn, _ := z.TagName()
                                candidateText = string(tn) == "td"
                        }
                }
        }
        return res
}

func FetchWeatherFromUrl(url string) []string {
        contents := uhtml.DownloadPage(url)
        log.Printf("content is %s", contents)
        return ParseWeather(bytes.NewReader(contents))
}

func RainWatcher(url string, channel string, responseChannel chan fmt.Stringer) {
        for {
                res := FetchWeatherFromUrl(url)
                log.Printf("Fetched %q from %s", res, url)
                if hasRain(res) {
                        log.Printf("Got rain, sending to chan", res, url)
                        responseChannel<- message.MsgSend{channel, strings.Join(res, "|")}
                }
                <-time.After(time.Minute * 30)
        }
}