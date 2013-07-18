package bsmeter

import (
        "net/http"
        "log"
        "regexp"
        html "code.google.com/p/go.net/html"
        "strings"
        "math"
        "utils/utilint"
        "encoding/json"
        "os"
        "fmt"
        "irc/message"
)

const scaleGood = 2
const minOcc = 5
const minProba = 0.1
const maxProba = 0.9
const defaultProba = 0.4

type BsState struct {
        GoodWords map[string] int
        BadWords map[string] int
        BsProba map[string] float64
        GoodUrls []string
        BadUrls []string
}

func defaultBsState() *BsState {
    bsState := new(BsState)
    bsState.GoodWords = map[string]int{}
    bsState.BadWords = map[string]int{}
    bsState.BsProba = map[string]float64{}
    bsState.GoodUrls = []string{}
    bsState.BadUrls = []string{}
    return bsState
}

type TrainingType int

type BsQuery struct {
        Urls []string
        IsTraining bool
        Bs bool
        Channel string
}

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

func TokenizePage(url string) []string {
        res := []string{}
        resp, err := http.Get(url)
        if err != nil {
                log.Printf("Got error : %v\n", err)
                return []string{}
        }
        defer resp.Body.Close()
        z := html.NewTokenizer(resp.Body)
        loop:
        for {
                tt := z.Next()
                switch tt {
                case html.ErrorToken:
                        break loop
                case html.TextToken:
                        text := strings.ToLower(string(z.Text()))
                        text = strings.Replace(text, "\n", "", -1)
                        for _, word := range strings.Split(text, " ") {
                                if strings.TrimSpace(word) != "" {
                                        res = append(res, word)
                                }
                        }
                }
        }
        return res
}

func enlargeCorpus(words []string, query BsQuery, state *BsState) {
        log.Printf("Adding %d words to corpus", len(words))
        if query.Bs {
                state.BadUrls = append(state.BadUrls, query.Urls...)
                for _, word := range words {
                    state.BadWords[word] = state.BadWords[word] + 1
                }
        } else {
                state.GoodUrls = append(state.GoodUrls, query.Urls...)
                for _, word := range words {
                    state.GoodWords[word] = state.GoodWords[word] + 1
                }
        }
        state.buildProba()
        saveBsState(state)
}

func (state *BsState) computeProbaForWord(word string) {
        occGood := state.GoodWords[word]
        occGood = occGood * scaleGood
        occBad := state.BadWords[word]
        if occGood + occBad < minOcc {
                return
        }
        propGood := math.Min(1, float64(occGood) / float64(len(state.GoodWords)))
        propBad := math.Min(1, float64(occBad) / float64(len(state.BadWords)))
        proba := propBad / (propBad + propGood)
        proba = math.Min(maxProba, proba)
        proba = math.Max(minProba, proba)
        state.BsProba[word] = proba
}

func (state *BsState) buildProba() {
        state.BsProba = map[string]float64{}
        for word := range state.GoodWords { state.computeProbaForWord(word) }
        for word := range state.BadWords { state.computeProbaForWord(word) }
}

type prob struct {
        word string
        proba float64
}

type probs []prob

func (p probs) Len() int {
        return len(p)
}

func (p probs) Less(i, j int) bool {
        dist1 := math.Abs(p[i].proba - 0.5)
        dist2 := math.Abs(p[j].proba - 0.5)
        return dist1 < dist2
}

func (p probs) Swap(i, j int) {
        p[i], p[j] = p[j], p[i]
}

func (p probs) Combined() float64 {
        num := 1.0
        denum := 1.0
        for _, prob := range p {
                num *= prob.proba
                denum *= (1 - prob.proba)
        }
        return num / (num + denum)
}

func (state *BsState) EvaluateBs(words []string) float64 {
        prbs := probs{}
        for _, word := range words {
                proba, found :=  state.BsProba[word]
                if !found {
                        proba = 0.4
                }
                prbs = append(prbs, prob{word, proba})
        }
        prbs = prbs[:utilint.MinInt(15, len(prbs))]
        log.Printf("Most significant probas are %v\n", prbs)
        return prbs.Combined()
}

func saveBsState(state *BsState) {
        file, _ := os.OpenFile("bstate", os.O_WRONLY | os.O_CREATE, 400)
        enc := json.NewEncoder(file)
        enc.Encode(state)
        file.Close()
}

func loadBsState() *BsState {
        file, err := os.Open("bstate")
        bsState := defaultBsState()
        if err != nil {
                return bsState
        }
        defer file.Close()
        dec := json.NewDecoder(file)
        dec.Decode(bsState)
        return bsState
}

func processQuery(query BsQuery, bsState *BsState) (string, bool) {
    log.Printf("Got bs query %v\n", query)
    results := []string{}
    for _, url := range query.Urls {
        title, found := LookupTitle(url)
        if !found {continue}
        words := TokenizePage(url)
        if query.IsTraining {
            enlargeCorpus(words, query, bsState)
            results = append(results, url)
        } else {
            score := bsState.EvaluateBs(words)
            results = append(results,
            fmt.Sprintf("[ %s : %.2f]", title, score))
        }
    }
    if len(results) == 0 { return "", false}
    if query.IsTraining {
        return fmt.Sprintf("Training with %s", strings.Join(results, " ")), true
    } else {
        return strings.Join(results, " "), true
    }
}

func BsWorker(requestChan chan BsQuery, responseChannel chan fmt.Stringer) {
    bsState := loadBsState()
    for {
        query := <-requestChan
        res, found := processQuery(query, bsState)
        if found {
            responseChannel<- message.MsgSend{query.Channel, res}
        }
    }
}
