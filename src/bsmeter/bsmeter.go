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
        "bytes"
        "os/exec"
        "utils/html"
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

func isPdf(url string) bool { return strings.HasSuffix(url, ".pdf") }

func savePdfToText(url *url.URL, content []byte, dir string) string {
        dest := path.Join(dir, url.Host, url.Path)
        tempDest := path.Join("/tmp", url.Host, url.Path)
        os.MkdirAll(filepath.Dir(tempDest), 500)
        os.MkdirAll(filepath.Dir(dest), 500)
        file, err := os.OpenFile(tempDest, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 400)
        if err != nil {
                log.Printf("Error on writing url %s, err: %s", tempDest, err)
                return ""
        }
        log.Printf("Saving %s in %s\n", url, tempDest)
        file.Write(content)
        file.Close()
        cmd := exec.Command("pdftotext", tempDest, dest)
        err = cmd.Run()
        if err != nil {
                log.Printf("Error on command %s : %s", cmd, err)
                return ""
        }
        return dest
}

func saveUrl(url *url.URL, content []byte, dir string) {
        dest := path.Join(dir, url.Host, url.Path)
        os.MkdirAll(filepath.Dir(dest), 500)
        file, err := os.OpenFile(dest, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 400)
        if err != nil {
                log.Printf("Error on writing url %s, err: %s", dest, err)
                return
        }
        log.Printf("Saving %s in %s\n", url, dest)
        file.Write(content)
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

func (bsState *BsState) trainWithPageContent(content []byte, bs bool) {
        words, _ := html.TokenizePage(bytes.NewReader(content))
        bsState.enlargeCorpus(words, bs)
}

func (bsState *BsState) trainWithPhrase(phrase string, bs bool) {
        words := html.TokenizeWords(phrase)
        bsState.enlargeCorpus(words, bs)
}

func (bsState *BsState) trainWithTextFile(path string, bs bool) {
        file, err := os.Open(path)
        if err != nil {
                log.Printf("Error on opening file %s", path)
                return
        }
        content, _ := ioutil.ReadAll(file)
        for _, line := range strings.Split(string(content), "\n") {
                bsState.trainWithPhrase(line ,bs)
        }
}

func (bsState *BsState) evaluateHtmlReader(url string, r io.Reader) BsResult{
        words, title := html.TokenizePage(r)
        score := bsState.EvaluateBs(words)
        return BsResult{title, score}
}

func (bsState *BsState) evaluateUrl(url string) BsResult{
        pageContent := html.DownloadPage(url)
        return bsState.evaluateHtmlReader(url,
                bytes.NewReader(pageContent))
}

func (bsState *BsState) evaluatePdf(strUrl string) BsResult{
        parsedUrl, _ := url.Parse(strUrl)
        pageContent := html.DownloadPage(strUrl)
        textFile := savePdfToText(parsedUrl,
                pageContent, "/tmp/temp_pdf")

        file, err := os.Open(textFile)
        if err != nil {
                log.Printf("Error on opening file %s", textFile)
                return BsResult{}
        }
        content, _ := ioutil.ReadAll(file)
        bsResult := bsState.evaluatePhrase(string(content))
        bsResult.Title = strUrl
        return bsResult
}

func truncatePhrase(phrase string, max int) string {
        if len(phrase) < max { return phrase }
        return fmt.Sprintf("%s...", phrase[:max])
}

func (bsState *BsState) evaluatePhrase(phrase string) BsResult{
        words := html.TokenizeWords(phrase)
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
        words, _ := html.TokenizePage(file)
        bsState.enlargeCorpus(words, bs)
}

func (bsState *BsState) processTrainQuery(query BsQuery) []string {
        storage := urlStorage[query.Bs]
        processed := []string{}
        for _, strUrl := range query.Urls {
                pageContent := html.DownloadPage(strUrl)
                parsedUrl, urlErr := url.Parse(strUrl)
                if urlErr != nil {
                        log.Printf("Invalid url, err: %s", urlErr)
                        continue
                }
                processed = append(processed, strUrl)
                if isPdf(strUrl) {
                        textPdf := savePdfToText(parsedUrl, pageContent, storage)
                        bsState.trainWithTextFile(textPdf, query.Bs)
                } else {
                        saveUrl(parsedUrl, pageContent, storage)
                        bsState.trainWithPageContent(pageContent, query.Bs)
                }
        }
        if query.Phrase != "" {
                dest := getPhraseStorage(query.Bs)
                appendPhrase(query.Phrase, dest)
                bsState.trainWithPhrase(query.Phrase, query.Bs)
        }
        return processed
}

func (bsState *BsState) evaluateQuery(query BsQuery) BsResults {
        results := []BsResult{}
        for _, url := range query.Urls {
                if isPdf(url) {
                        results = append(results, bsState.evaluatePdf(url))
                } else {
                        bsResult := bsState.evaluateUrl(url)
                        if bsResult.Title != "" {
                            results = append(results, bsResult)
                        }
                }
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
                if info.IsDir() {
                        return nil
                }
                if isPdf(path) {
                        log.Printf("Loading pdf path %s with bs = %v", path, bs)
                        bsState.trainWithTextFile(path, bs)
                } else {
                        log.Printf("Loading path %s with bs = %v", path, bs)
                        bsState.trainWithHtmlFile(path, bs)
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
        bsState.trainWithTextFile(getPhraseStorage(true), true)
        bsState.trainWithTextFile(getPhraseStorage(false), false)
}

func BsWorker(requestChan chan BsQuery, responseChannel chan fmt.Stringer) {
        bsState := loadBsState(bsFile)
        for {
                query := <-requestChan
                if query.IsTraining {
                        processedUrls := bsState.processTrainQuery(query)
                        bsState.save(bsFile)
                        if len(processedUrls) > 0 {
                            responseChannel<- message.MsgSend{query.Channel, fmt.Sprintf("Training with %s", strings.Join(processedUrls, " "))}
                        }
                } else if query.IsReload {
                        bsState.processReload()
                        bsState.save(bsFile)
                } else {
                        results := bsState.evaluateQuery(query)
                        if len(results) > 0 {
                            responseChannel<- message.MsgSend{query.Channel, results.String()}
                        }
                }
        }
}
