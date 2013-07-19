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
)

const bsFile = "bsState"
var urlStorage = map[bool]string{false : "good", true: "bad"}

type BsQuery struct {
        Urls []string
        IsTraining bool
        Bs bool
        Channel string
        IsReload bool
}

type BsResults []BsResult
type BsResult struct {
        Url string
        Title string
        Score float64
}

func (res BsResults) String() string {
        resStr := []string{}
        for _, v := range res {
                resStr = append(resStr, v.String())
        }
        return strings.Join(resStr, " ")
}

func (res BsResult) String() string {
        return fmt.Sprintf("[ %s : %.2f]", res.Title, res.Score)
}

func saveUrl(strUrl, content, dir string) string {
        parsedUrl, urlErr := url.Parse(strUrl)
        if urlErr != nil {
                log.Printf("Invalid url, err: %s", urlErr)
                return ""
        }
        dest := path.Join(dir, parsedUrl.Host, parsedUrl.Path)
        os.MkdirAll(filepath.Dir(dest), 500)
        file, err := os.OpenFile(dest, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 400)
        if err != nil {
                log.Printf("Error on writing url %s, err: %s", dest, err)
                return ""
        }
        log.Printf("Saving %s in %s\n", strUrl, dest)
        file.WriteString(content)
        file.Close()
        return content
}

func (bsState *BsState) trainWithContent(content string, bs bool) {
        words, _ := TokenizePage(strings.NewReader(content))
        bsState.enlargeCorpus(words, bs)
}

func (bsState *BsState) evaluateReader(url string, r io.Reader) BsResult{
        words, title := TokenizePage(r)
        score := bsState.EvaluateBs(words)
        return BsResult{url, title, score}
}

func (bsState *BsState) evaluateContent(url, content string) BsResult{
        return bsState.evaluateReader(url, strings.NewReader(content))
}

func (bsState *BsState) evaluateUrl(url string) BsResult{
        pageContent := downloadPage(url)
        return bsState.evaluateContent(url, pageContent)
}

func (bsState *BsState) evaluateFile(filename string) BsResult{
        file, err := os.Open(filename)
        if err != nil {
                log.Printf("Error on opening file %s", err)
                return BsResult{}
        }
        defer file.Close()
        return bsState.evaluateReader(filename, file)
}

func (bsState *BsState) trainWithFile(filename string, bs bool) {
        file, err := os.Open(filename)
        if err != nil {
                log.Printf("Error on opening file %s", err)
                return
        }
        words, _ := TokenizePage(file)
        bsState.enlargeCorpus(words, bs)
}

func (bsState *BsState) processTrainQuery(query BsQuery) {
        contents := []string{}
        for _, url := range query.Urls {
                pageContent := downloadPage(url)
                content := saveUrl(url, pageContent, urlStorage[query.Bs])
                contents = append(contents, content)
                bsState.trainWithContent(content, query.Bs)
        }
}

func (bsState *BsState) processEvaluateQuery(query BsQuery) BsResults {
        results := []BsResult{}
        for _, url := range query.Urls {
                results = append(results, bsState.evaluateUrl(url))
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
                        bsState.trainWithFile(path, bs)
                }
                return nil
        }
}

func (bsState *BsState) processReload() {
        bsState.GoodWords = map[string]int {}
        bsState.BadWords = map[string]int {}
        bsState.BsProba = map[string]float64 {}
        filepath.Walk(urlStorage[true], bsState.walker(true))
        filepath.Walk(urlStorage[false], bsState.walker(false))
}

func BsWorker(requestChan chan BsQuery, responseChannel chan fmt.Stringer) {
        bsState := loadBsState(bsFile)
        for {
                query := <-requestChan
                if query.IsTraining {
                        bsState.processTrainQuery(query)
                        bsState.save(bsFile)
                        responseChannel<- message.MsgSend{query.Channel,
                                fmt.Sprintf("Training with urls %v", query.Urls)}
                } else if query.IsReload {
                        bsState.processReload()
                        bsState.save(bsFile)
                } else {
                        results := bsState.processEvaluateQuery(query)
                        responseChannel<- message.MsgSend{query.Channel, results.String()}
                }
        }
}
