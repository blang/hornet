package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type PubError struct {
	Error string `json:"error"`
}

type PubMessage struct {
	Msg string `json:"msg"`
}

type RestAPI struct {
	router  *http.ServeMux
	service *Service
}

func NewRestAPI(service *Service) *RestAPI {
	api := &RestAPI{
		router:  http.NewServeMux(),
		service: service,
	}
	api.RegisterEndPoints()
	return api
}

func (api *RestAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin,Authorization,Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
	} else {
		api.router.ServeHTTP(w, r)
	}
}

func (api *RestAPI) RegisterEndPoints() {
	api.router.HandleFunc("/config", api.handleConfig)

}

type PubConfig struct {
	Arma2OAPath  string `json:"arma2oapath"`
	Arma2Path    string `json:"arma2path"`
	Arma3Path    string `json:"arma3path"`
	Arma2CO      bool   `json:"arma2co"`
	Arma2Profile string `json:"arma2profile"`
	Arma3Profile string `json:"arma3profile"`
}

func (api *RestAPI) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		api.handleConfigGet(w, r)
	case "POST":
		api.handleConfigPost(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}

}

func (api *RestAPI) handleConfigGet(w http.ResponseWriter, r *http.Request) {
	pubConf := api.buildPubConfig(api.service.Config())
	b, err := json.Marshal(pubConf)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func (api *RestAPI) handleConfigPost(w http.ResponseWriter, r *http.Request) {
	var pubConf PubConfig
	err := json.NewDecoder(r.Body).Decode(&pubConf)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Save Config
	err = api.savePubConfig(api.service.Config(), &pubConf)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	api.handleConfigGet(w, r)
}

func (api *RestAPI) buildPubConfig(config *Config) *PubConfig {
	if config == nil {
		return &PubConfig{}
	}
	return &PubConfig{
		Arma2OAPath:  config.Arma2OAPath,
		Arma2Path:    config.Arma2Path,
		Arma3Path:    config.Arma3Path,
		Arma2CO:      config.Arma2CO,
		Arma2Profile: config.Arma2Profile,
		Arma3Profile: config.Arma3Profile,
	}
}

func (api *RestAPI) savePubConfig(config *Config, pubConf *PubConfig) error {
	if config == nil || pubConf == nil {
		return fmt.Errorf("Could not save config")
	}
	config.Arma2OAPath = pubConf.Arma2OAPath
	config.Arma2Path = pubConf.Arma2Path
	config.Arma2Profile = pubConf.Arma2Profile
	config.Arma3Path = pubConf.Arma3Path
	config.Arma3Profile = pubConf.Arma3Profile
	config.Arma2CO = pubConf.Arma2CO
	return nil
}
