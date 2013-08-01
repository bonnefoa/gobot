package bsmeter

import (
	"bytes"
	"encoding/json"
	bstrings "github.com/bonnefoa/gobot/utils/strings"
	"log"
	"math"
	"os"
	"github.com/bonnefoa/gobot/utils/html"
	"net/url"
	"path/filepath"
	"io"
	"io/ioutil"
        "strings"
        "errors"
        "fmt"
)

type BsState struct {
	GoodWords map[string]int
	BadWords  map[string]int
	BsProba   map[string]float64
        BsConf
}

func defaultBsState() *BsState {
        bsState := new(BsState)
        bsState.GoodWords = map[string]int{}
        bsState.BadWords = map[string]int{}
        bsState.BsProba = map[string]float64{}
        bsState.BsConf = BsConf{}
        return bsState
}


func (s *BsState) save() {
	file, _ := os.OpenFile(s.StateFile,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 400)
	enc := json.NewEncoder(file)
	enc.Encode(s)
	file.Close()
}

func (state *BsState) enlargeCorpus(words []string, isBs bool) {
	log.Printf("Adding %d words to corpus", len(words))
	if isBs {
		for _, word := range words {
			state.BadWords[word] = state.BadWords[word] + 1
		}
	} else {
		for _, word := range words {
			state.GoodWords[word] = state.GoodWords[word] + 1
		}
	}
	state.rebuildProbaMap()
}

func (state *BsState) computeProbaForWord(word string) {
	occGood := state.GoodWords[word]
	occGood = occGood * scaleGood
	occBad := state.BadWords[word]
	if occGood+occBad < minOcc {
		return
	}
	propGood := math.Min(1, float64(occGood)/math.Max(1, float64(len(state.GoodWords))))
	propBad := math.Min(1, float64(occBad)/math.Max(1, float64(len(state.BadWords))))
	proba := propBad / (propBad + propGood)
	proba = math.Min(maxProba, proba)
	proba = math.Max(minProba, proba)
	state.BsProba[word] = proba
}

func (state *BsState) rebuildProbaMap() {
	state.BsProba = map[string]float64{}
	for word := range state.GoodWords {
		state.computeProbaForWord(word)
	}
	for word := range state.BadWords {
		state.computeProbaForWord(word)
	}
}

func (state *BsState) EvaluateBs(words []string) float64 {
	prbs := probs{}
	for _, word := range words {
		proba, found := state.BsProba[word]
		if !found {
			proba = 0.4
		}
		prbs = append(prbs, prob{word, proba})
	}
        prbs = prbs.MostSignificantProbas(15)
	log.Printf("Most significant probas are %v\n", prbs)
	return prbs.Combined()
}

func (bsState *BsState) evaluatePhrase(v interface{}) (*BsResult, error) {
        var phrase string
        switch tv := v.(type) {
        case io.Reader :
                bytes, err := ioutil.ReadAll(tv)
                if err != nil { return nil, err }
                phrase = string(bytes)
        case string:
                phrase = tv
        default:
                return nil, errors.New(fmt.Sprintf("Unkown type v, %q", v))
        }
	words := bstrings.TokenizeWords(phrase)
	return &BsResult{bstrings.TruncatePhrase(phrase, 10), bsState.EvaluateBs(words)}, nil
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

func (bsState *BsState) processTrainQuery(query BsQuery) ([]string, error) {
	storage := bsState.getStorage(query.Bs)
	processed := []string{}
	for _, strUrl := range query.Urls {
		pageContent, err := html.DownloadPage(strUrl)
                if err != nil { return []string{}, err }
		parsedUrl, err := url.Parse(strUrl)
                if err != nil { return []string{}, err }
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
		dest := bsState.getPhraseStorage(query.Bs)
		appendPhrase(query.Phrase, dest)
		bsState.trainWithPhrase(query.Phrase, query.Bs)
	}
	return processed, nil
}

func (bsState *BsState) walker(bs bool) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
                log.Printf("Walking in %s\n", path)
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

func (bsState *BsState) trainWithPageContent(content []byte, bs bool) {
	words, _ := html.TokenizePage(bytes.NewReader(content))
	bsState.enlargeCorpus(words, bs)
}

func (bsState *BsState) trainWithPhrase(phrase string, bs bool) {
	words := bstrings.TokenizeWords(phrase)
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
		bsState.trainWithPhrase(line, bs)
	}
}

func (bsState *BsState) evaluateHtml(strUrl string, v interface{}) (*BsResult, error) {
        switch tv := v.(type) {
        case io.Reader:
                words, title := html.TokenizePage(tv)
                score := bsState.EvaluateBs(words)
                return &BsResult{title, score}, nil
        case string:
                parsedUrl, err := url.Parse(tv)
                if err != nil { return nil, err }
                var pageContent []byte
                if parsedUrl.Scheme == "http" {
                        pageContent, err = html.DownloadPage(tv)
                } else {
                        pageContent, err = ioutil.ReadFile(tv)
                }
                if err != nil { return nil, err }
                r := bytes.NewReader(pageContent)
                return bsState.evaluateHtml(strUrl, r)
        default:
                return nil, errors.New(fmt.Sprintf("Unhandled parameter %q", v))
        }
}

func (bsState *BsState) evaluatePdf(strUrl string) (*BsResult, error) {
        parsedUrl, err := url.Parse(strUrl)
        pdfFile := strUrl
        if parsedUrl.Scheme == "http" {
                content, err := html.DownloadPage(parsedUrl.String())
                if err != nil { return nil, err }
                pdfFile = savePdfToText(parsedUrl, content, "/tmp/temp_pdf")
        }
        r, err := os.Open(pdfFile)
        if err != nil { return nil, err }
        return bsState.evaluatePhrase(r)
}

func (bsState *BsState) evaluateUrl(strUrl string) (*BsResult, error) {
        if isPdf(strUrl) {
                return bsState.evaluatePdf(strUrl)
        }
        return bsState.evaluateHtml(strUrl, strUrl)
}

func (bsState *BsState) evaluateQuery(query BsQuery) BsResults {
	results := []*BsResult{}
	for _, strUrl := range query.Urls {
                bsResult, err := bsState.evaluateUrl(strUrl)
                if err != nil {
                        log.Printf("Got err for evaluation of %s : %s", strUrl, err)
                        continue
                }
                results = append(results, bsResult)
                if bsResult.Title != "" {
                        results = append(results, bsResult)
                }
	}
	if query.Phrase != "" {
                res, err := bsState.evaluatePhrase(query.Phrase)
                if err != nil {
                        log.Printf("Got err for evaluation of %s : %s", query.Phrase, err)
                } else {
                        results = append(results, res)
                }
	}
	return results
}

func (bsState *BsState) processReload() {
	bsState.GoodWords = map[string]int{}
	bsState.BadWords = map[string]int{}
	bsState.BsProba = map[string]float64{}
	filepath.Walk(bsState.getStorage(true), bsState.walker(true))
	filepath.Walk(bsState.getStorage(false), bsState.walker(false))
	bsState.trainWithTextFile(bsState.getPhraseStorage(true), true)
	bsState.trainWithTextFile(bsState.getPhraseStorage(false), false)
}
