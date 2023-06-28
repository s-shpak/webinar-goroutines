package common

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

//go:embed first.names.json
var firstNamesContents []byte

//go:embed last.names.json
var lastNamesContents []byte

type NamesFetcher struct {
	firstNames []string
	lastNames  []string
	mux        *sync.Mutex
	r          *rand.Rand
}

var namesFetcher *NamesFetcher

func NewNamesFetcher() (*NamesFetcher, error) {
	var err error
	if namesFetcher == nil {
		namesFetcher, err = newNamesFetcher()
		if err != nil {
			return nil, err
		}
	}
	return namesFetcher, nil
}

func newNamesFetcher() (*NamesFetcher, error) {
	f := &NamesFetcher{
		r:   rand.New(rand.NewSource(time.Now().Unix())),
		mux: &sync.Mutex{},
	}
	if err := json.Unmarshal(firstNamesContents, &f.firstNames); err != nil {
		return nil, fmt.Errorf("failed to unmarshal first names file: %w", err)
	}
	if err := json.Unmarshal(lastNamesContents, &f.lastNames); err != nil {
		return nil, fmt.Errorf("failed to unmarshal last names file: %w", err)
	}
	return f, nil
}

type Name struct {
	FirstName string
	LastName  string
}

func (f *NamesFetcher) GetNames(dst []Name) {
	for i := range dst {
		dst[i] = f.chooseRandName()
	}
}

func (f *NamesFetcher) chooseRandName() Name {
	f.mux.Lock()
	defer f.mux.Unlock()
	firstNameIdx, lastNameIdx := f.r.Intn(len(f.firstNames)), f.r.Intn(len(f.lastNames))
	return Name{
		FirstName: f.firstNames[firstNameIdx],
		LastName:  f.lastNames[lastNameIdx],
	}
}
