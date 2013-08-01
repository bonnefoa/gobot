package bsmeter

import (
	"fmt"
	"os"
	"encoding/json"
)

type BsConf struct {
        StateFile string
        StorageGood string
        StorageBad string
}

func (c *BsConf) loadBsState() *BsState {
	file, err := os.Open(c.StateFile)
	bsState := &BsState{}
	if err != nil {
		return bsState
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	dec.Decode(bsState)
        bsState.BsConf = *c
	return bsState
}

func (c *BsConf) getStorage(bs bool) string {
        storage := c.StorageGood
        if bs {
                storage = c.StorageBad
        }
	return storage
}

func (c *BsConf) getPhraseStorage(bs bool) string {
	return fmt.Sprintf("%s_file", c.getStorage(bs))
}
