package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
)

type Store struct {
	Config Config
}

type Config struct {
	Arma2OAPath  string
	Arma2Path    string
	Arma3Path    string
	Arma2CO      bool
	Arma2Profile string
	Arma3Profile string
}

func NewStore() *Store {
	return &Store{}
}

func RestoreStore(filename string) (*Store, error) {
	store, err := restoreStore(filename)
	if err != nil {
		store = NewStore()
	}
	return store, err
}

func restoreStore(filename string) (*Store, error) {
	if filename == "" {
		return nil, errors.New("Store file could not be empty")
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return nil, fmt.Errorf("Could not open store file %q: %q", filename, err)
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	var store Store
	err = dec.Decode(&store)
	if err != nil {
		return nil, fmt.Errorf("Could not decode store from file %q: %q", filename, err)
	}
	return &store, nil
}

func (s *Store) SaveStore(filename string) error {
	if filename == "" {
		return errors.New("Store file could not be empty")
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("Could not open store file %q: %q", filename, err)
	}
	defer f.Close()
	dec := gob.NewEncoder(f)
	err = dec.Encode(s)
	if err != nil {
		return fmt.Errorf("Could not encode store to file %q: %q", filename, err)
	}
	return nil
}
