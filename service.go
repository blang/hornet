package main

import (
	"errors"
	"fmt"
	"github.com/blang/procd"
	"os"
	"path"
	"strconv"
	"strings"
)

type LaunchConfig struct {
	Name      string `json:"name"`
	Game      string `json:"game"`
	Modstring string `json:"modstring"`
	Betamod   string `json:"betamod"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Password  string `json:"password"`
}

type Service struct {
	store   *Store
	process *procd.Process
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Config() *Config {
	return &(s.store.Config)
}

func (s *Service) Launch(conf *LaunchConfig) error {
	if s.process != nil && s.process.State() != procd.RunStateStopped {
		return errors.New("Instance already running")
	}
	switch conf.Game {
	case "a2":
		err := s.checkArma2OA()
		if err != nil {
			return err
		}
		cwd, exe, args := s.compileParams(s.Config().Arma2OAPath, "arma2oa.exe", s.Config().Arma2Params, s.Config().Arma2Profile, conf)
		s.process = procd.NewProcess(cwd, exe, args)
		s.process.Start()
	case "a3":
		err := s.checkArma3()
		if err != nil {
			return err
		}
		cwd, exe, args := s.compileParams(s.Config().Arma3Path, "arma3.exe", s.Config().Arma3Params, s.Config().Arma3Profile, conf)
		s.process = procd.NewProcess(cwd, exe, args)
		s.process.Start()
		return errors.New("Not yet implemented")
	default:
		return errors.New("Unknown gametype")
	}
	return nil
}

func (s *Service) Shutdown() {
	if s.process != nil && s.process.State() != procd.RunStateStopped {
		s.process.Stop()
		s.process.Wait()
	}
}

func (s *Service) checkArma2OA() error {
	if s.Config().Arma2OAPath == "" {
		return errors.New("Arma2OAPath not configured")
	}
	fi, err := os.Stat(s.Config().Arma2OAPath)
	if err != nil {
		return fmt.Errorf("Arma2OA Path not exists: %q", s.Config().Arma2OAPath)
	}
	if !fi.IsDir() {
		return fmt.Errorf("Arma2OA Path is not a directory: %q", s.Config().Arma2OAPath)
	}
	exePath := path.Join(s.Config().Arma2OAPath, "arma2oa.exe")
	fi, err = os.Stat(exePath)
	if err != nil {
		return fmt.Errorf("Arma2OA Executable path not exists: %q", exePath)
	}
	if !(fi.Size() > 0) {
		return fmt.Errorf("Arma2OA Executable not exists: ", exePath)
	}

	return nil
}

func (s *Service) checkArma3() error {
	if s.Config().Arma3Path == "" {
		return errors.New("Arma3Path not configured")
	}
	fi, err := os.Stat(s.Config().Arma3Path)
	if err != nil {
		return fmt.Errorf("Arma3 Path not exists: %q", s.Config().Arma3Path)
	}
	if !fi.IsDir() {
		return fmt.Errorf("Arma3 Path is not a directory: %q", s.Config().Arma3Path)
	}
	exePath := path.Join(s.Config().Arma3Path, "arma3.exe")
	fi, err = os.Stat(exePath)
	if err != nil {
		return fmt.Errorf("Arma3 Executable path not exists: %q", exePath)
	}
	if !(fi.Size() > 0) {
		return fmt.Errorf("Arma3 Executable not exists: ", exePath)
	}

	return nil
}

func (s *Service) compileParams(cwd, exe string, addParams string, profile string, conf *LaunchConfig) (string, string, []string) {
	startPath := path.Join(cwd, exe)
	if conf.Betamod != "" {
		startPath = path.Join(cwd, conf.Betamod, exe)
	}

	args := []string{
		"-nosplash",
		"-world=empty",
		"-pause",
		"-connect=" + conf.Host,
		"-port=" + strconv.Itoa(conf.Port),
		"-password=" + conf.Password,
		"-profiles=" + profile,
		"-name=" + profile,
		"-mod=" + conf.Modstring,
	}
	args = append(args, strings.Split(addParams, " ")...)
	return cwd, startPath, args
}
