package main

import (
	"encoding/json"
	"fmt"
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
	api.router.HandleFunc("/launch", api.handleLaunch)

}

type PubConfig struct {
	Arma2OAPath  string `json:"arma2oapath"`
	Arma2Profile string `json:"arma2profile"`
	Arma2Params  string `json:"arma2params"`
	Arma3Path    string `json:"arma3path"`
	Arma3Profile string `json:"arma3profile"`
	Arma3Params  string `json:"arma3params"`
}

type PubLaunchConfig struct {
	Name      string `json:"name"`
	Game      string `json:"game"`
	Modstring string `json:"modstring"`
	Betamod   string `json:"betamod"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Password  string `json:"password"`
}

func (c *PubLaunchConfig) LaunchConfig() *LaunchConfig {
	return &LaunchConfig{
		Name:      c.Name,
		Game:      c.Game,
		Modstring: c.Modstring,
		Betamod:   c.Betamod,
		Host:      c.Host,
		Port:      c.Port,
		Password:  c.Password,
	}
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
		Arma2Profile: config.Arma2Profile,
		Arma2Params:  config.Arma2Params,
		Arma3Path:    config.Arma3Path,
		Arma3Profile: config.Arma3Profile,
		Arma3Params:  config.Arma3Params,
	}
}

func (api *RestAPI) savePubConfig(config *Config, pubConf *PubConfig) error {
	if config == nil || pubConf == nil {
		return fmt.Errorf("Could not save config")
	}
	config.Arma2OAPath = pubConf.Arma2OAPath
	config.Arma2Profile = pubConf.Arma2Profile
	config.Arma2Params = pubConf.Arma2Params
	config.Arma3Path = pubConf.Arma3Path
	config.Arma3Profile = pubConf.Arma3Profile
	config.Arma3Params = pubConf.Arma3Params
	return nil
}

func (api *RestAPI) handleLaunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var pubLaunchConfig PubLaunchConfig
	err := json.NewDecoder(r.Body).Decode(&pubLaunchConfig)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	launchConfig := pubLaunchConfig.LaunchConfig()
	err = api.service.Launch(launchConfig)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		b, _ := json.Marshal(&PubError{err.Error()})
		w.Write(b)
	}
}
