package meteo

import (
	"bytes"
	html "code.google.com/p/go.net/html"
	"fmt"
	"github.com/bonnefoa/gobot/message"
	uhtml "github.com/bonnefoa/gobot/utils/html"
	"io"
	"log"
	"regexp"
	"strings"
	"time"
)

type Meteo struct {
	Url     string
	Channel string
	Period  int // In minutes
}

var reRain, _ = regexp.Compile("Pluie [fm]")

func hasRain(parsed []string) bool {
	for _, el := range parsed {
		if reRain.MatchString(el) {
			return true
		}
	}
	return false
}

func hasTablPluieClass(z *html.Tokenizer) bool {
	key, val, more := z.TagAttr()
	if string(key) == "class" && string(val) == "tablPluie" {
		return true
	}
	if more {
		return hasTablPluieClass(z)
	}
	return false
}

func ParseWeather(r io.Reader) []string {
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

func FetchWeatherFromUrl(url string) ([]string, error) {
	contents, err := uhtml.DownloadPage(url)
        if err != nil { return []string{}, err }
	return ParseWeather(bytes.NewReader(contents)), nil
}

func RainWatcher(meteo Meteo, responseChannel chan fmt.Stringer) {
	for {
		res, err := FetchWeatherFromUrl(meteo.Url)
                if err != nil {
                        log.Printf("Got error on fetch : %s", res)
                        continue
                }
		log.Printf("Fetched %q from %s", res, meteo.Url)
		if hasRain(res) {
			log.Printf("Got rain, sending to chan", res, meteo.Url)
			responseChannel <- message.MsgSend{meteo.Channel,
				strings.Join(res, "|")}
		}
		<-time.After(time.Duration(meteo.Period) * time.Minute)
	}
}
