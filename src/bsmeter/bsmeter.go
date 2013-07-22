package bsmeter

import (
        "log"
        "strings"
        "fmt"
        "irc/message"
        "path"
        "path/filepath"
        "os"
        "net/url"
        "io"
        "io/ioutil"
)

const bsFile = "bsState"
var urlStorage = map[bool]string{false : "good", true: "bad"}

type BsQuery struct {
        Phrase string
        Urls []string
        IsTraining bool
        Bs bool
        Channel string
        IsReload bool
}

type BsResults []BsResult
type BsResult struct {
        Title string
        Score float64
}

func (query BsQuery) String() string {
        if query.IsTraining {
                if len(query.Urls) > 0 {
                        return fmt.Sprintf("Training with urls %s", query.Urls)
                }
                return fmt.Sprintf("Training with phrase %s", query.Phrase)
        }
        return fmt.Sprintf("%v", query)
}

func (res BsResults) String() string {
        resStr := []string{}
        for _, v := range res {
                resStr = append(resStr, v.String())
        }
        return strings.Join(resStr, " ")
}

func getPhraseStorage(bs bool) string {
        return fmt.Sprintf("%s_file", urlStorage[bs])
}

func (res BsResult) String() string {
        return fmt.Sprintf("[ %s : %.2f]", res.Title, res.Score)
}

func saveUrl(strUrl, content, dir string) {
        parsedUrl, urlErr := url.Parse(strUrl)
        if urlErr != nil {
                log.Printf("Invalid url, err: %s", urlErr)
                return
        }
        dest := path.Join(dir, parsedUrl.Host, parsedUrl.Path)
        os.MkdirAll(filepath.Dir(dest), 500)
        file, err := os.OpenFile(dest, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 400)
        if err != nil {
                log.Printf("Error on writing url %s, err: %s", dest, err)
                return
        }
        log.Printf("Saving %s in %s\n", strUrl, dest)
        file.WriteString(content)
        file.Close()
}

func appendPhrase(phrase, dest string) {
        file, err := os.OpenFile(dest, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 400)
        if err != nil {
                log.Printf("Error on writing phrase %s, err: %s", phrase, err)
                return
        }
        file.WriteString("\n")
        file.WriteString(phrase)
        file.Close()
}

func (bsState *BsState) trainWithPageContent(content string, bs bool) {
        words, _ := TokenizePage(strings.NewReader(content))
        bsState.enlargeCorpus(words, bs)
}

func (bsState *BsState) trainWithPhrase(phrase string, bs bool) {
        words := tokenizeWords(phrase)
        bsState.enlargeCorpus(words, bs)
}

func (bsState *BsState) evaluateHtmlReader(url string, r io.Reader) BsResult{
        words, title := TokenizePage(r)
        score := bsState.EvaluateBs(words)
        return BsResult{title, score}
}

func (bsState *BsState) evaluateUrl(url string) BsResult{
        pageContent := downloadPage(url)
        return bsState.evaluateHtmlReader(url, strings.NewReader(pageContent))
}

func truncatePhrase(phrase string, max int) string {
        if len(phrase) < max { return phrase }
        return fmt.Sprintf("%s...", phrase[:max])
}

func (bsState *BsState) evaluatePhrase(phrase string) BsResult{
        words := tokenizeWords(phrase)
        return BsResult{truncatePhrase(phrase, 10), bsState.EvaluateBs(words)}
}

func (bsState *BsState) evaluateHtmlFile(filename string) BsResult{
        file, err := os.Open(filename)
        if err != nil {
                log.Printf("Error on opening file %s", err)
                return BsResult{}
        }
        defer file.Close()
        return bsState.evaluateHtmlReader(filename, file)
}

func (bsState *BsState) trainWithHtmlFile(filename string, bs bool) {
        file, err := os.Open(filename)
        if err != nil {
                log.Printf("Error on opening file %s", err)
                return
        }
        words, _ := TokenizePage(file)
        bsState.enlargeCorpus(words, bs)
}

func (bsState *BsState) processTrainQuery(query BsQuery) {
        storage := urlStorage[query.Bs]
        for _, url := range query.Urls {
                pageContent := downloadPage(url)
                saveUrl(url, pageContent, storage)
                bsState.trainWithPageContent(pageContent, query.Bs)
        }
        if query.Phrase != "" {
                dest := getPhraseStorage(query.Bs)
                appendPhrase(query.Phrase, dest)
                bsState.trainWithPhrase(query.Phrase, query.Bs)
        }
}

func (bsState *BsState) evaluateQuery(query BsQuery) BsResults {
        results := []BsResult{}
        for _, url := range query.Urls {
                results = append(results, bsState.evaluateUrl(url))
        }
        if query.Phrase != "" {
                results = append(results, bsState.evaluatePhrase(query.Phrase))
        }
        return results
}

func (bsState *BsState) walker(bs bool) func(path string, info os.FileInfo, err error) error {
        return func(path string, info os.FileInfo, err error) error {
                if err != nil {
                        log.Printf("Error on walk : %s\n", err)
                        return err
                }
                if !info.IsDir() {
                        log.Printf("Loading path %s with bs = %v", path, bs)
                        bsState.trainWithHtmlFile(path, bs)
                }
                return nil
        }
}

func (bsState *BsState) reloadPhrases(bs bool) {
        storage := getPhraseStorage(bs)
        file, err := os.Open(storage)
        if err != nil {
                log.Printf("Error on opening file %s", storage)
                return
        }
        content, _ := ioutil.ReadAll(file)
        for _, line := range strings.Split(string(content), "\n") {
                bsState.trainWithPhrase(line ,bs)
        }
}

func (bsState *BsState) processReload() {
        bsState.GoodWords = map[string]int {}
        bsState.BadWords = map[string]int {}
        bsState.BsProba = map[string]float64 {}
        filepath.Walk(urlStorage[true], bsState.walker(true))
        filepath.Walk(urlStorage[false], bsState.walker(false))
        bsState.reloadPhrases(true)
        bsState.reloadPhrases(false)
}

func BsWorker(requestChan chan BsQuery, responseChannel chan fmt.Stringer) {
        bsState := loadBsState(bsFile)
        for {
                query := <-requestChan
                if query.IsTraining {
                        bsState.processTrainQuery(query)
                        bsState.save(bsFile)
                        responseChannel<- message.MsgSend{query.Channel, query.String()}
                } else if query.IsReload {
                        bsState.processReload()
                        bsState.save(bsFile)
                } else {
                        results := bsState.evaluateQuery(query)
                        responseChannel<- message.MsgSend{query.Channel, results.String()}
                }
        }
}
