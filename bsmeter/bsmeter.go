package bsmeter

import (
	"fmt"
	"github.com/bonnefoa/gobot/message"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type BsQuery struct {
	Phrase     string
	Urls       []string
	IsTraining bool
	Bs         bool
	Channel    string
	IsReload   bool
}

type BsResults []*BsResult
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

func (res BsResult) String() string {
	return fmt.Sprintf("[ %s : %.2f]", res.Title, res.Score)
}

func isPdf(url string) bool { return strings.HasSuffix(url, ".pdf") }

func savePdfToText(url *url.URL, content []byte, dir string) string {
	dest := path.Join(dir, url.Host, url.Path)
	tempDest := path.Join("/tmp", url.Host, url.Path)
	os.MkdirAll(filepath.Dir(tempDest), 500)
	os.MkdirAll(filepath.Dir(dest), 500)
	file, err := os.OpenFile(tempDest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 400)
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
	file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 400)
	if err != nil {
		log.Printf("Error on writing url %s, err: %s", dest, err)
		return
	}
	log.Printf("Saving %s in %s\n", url, dest)
	file.Write(content)
	file.Close()
}

func appendPhrase(phrase, dest string) {
	file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 400)
	if err != nil {
		log.Printf("Error on writing phrase %s, err: %s", phrase, err)
		return
	}
	file.WriteString("\n")
	file.WriteString(phrase)
	file.Close()
}

func BsWorker(bsConf BsConf, requestChan chan BsQuery, responseChannel chan fmt.Stringer) {
	bsState := bsConf.loadBsState()
	for {
		query := <-requestChan
		if query.IsTraining {
			processedUrls, err := bsState.processTrainQuery(query)
                        if err != nil {
                                log.Printf("Error on training with %s, err: %s", query, err)
                                continue
                        }
			bsState.save()
			if len(processedUrls) > 0 {
				responseChannel <- message.MsgSend{query.Channel,
                                        fmt.Sprintf("Training with %s", strings.Join(processedUrls, " "))}
			}
		} else if query.IsReload {
			bsState.processReload()
			bsState.save()
		} else {
			results := bsState.evaluateQuery(query)
			if len(results) > 0 {
				responseChannel <- message.MsgSend{query.Channel, results.String()}
			}
		}
	}
}
