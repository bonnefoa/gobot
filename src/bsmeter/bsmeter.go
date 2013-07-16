package bsmeter

import (
        "net/http"
        "log"
        "encoding/xml"
        "regexp"
)

var reUrls, _ = regexp.Compile("https?://[^ ]*")

func ExtractUrls(mess string) []string {
        return reUrls.FindAllString(mess, -1)
}

func LookupTitle(url string) string {
        resp, err := http.Get(url)
        if err != nil {
                log.Printf("Got error is %v\n", err)
                return ""
        }
        dec := xml.NewDecoder(resp.Body)
        isTitle := false
        for {
                t, _ := dec.Token()
                if t == nil { break }
                switch token := t.(type) {
                case xml.StartElement:
                        if token.Name.Local == "title" {
                                isTitle = true
                        }
                case xml.EndElement:
                        if token.Name.Local == "title" {
                                isTitle = false
                                resp.Body.Close()
                                return ""
                        }
                case xml.CharData:
                        if isTitle {
                                resp.Body.Close()
                                return string(token)
                        }
                }
        }
        return ""
}
