package main

import "github.com/go-ini/ini"
import "github.com/tj/go-debug"
import "os"
import "fmt"
import "errors"

var badCfgErr = errors.New("Bad configuration file!")

var debugCfg = debug.Debug("oktad:config")

type OktaConfig struct {
	BaseURL string
	AppURL  string
}

// loads configuration data from the file specified
func parseConfig(fname string) (OktaConfig, error) {
	var cfg OktaConfig

	f, err := loadConfig(fname)

	if err != nil {
		return cfg, err
	}

	osec := f.Section("okta")
	if osec == nil {
		return cfg, badCfgErr
	}

	if !osec.HasKey("baseUrl") || !osec.HasKey("appUrl") {
		return cfg, badCfgErr
	}

	bu, err := osec.GetKey("baseUrl")
	if err != nil {
		return cfg, err
	}

	au, err := osec.GetKey("appUrl")
	if err != nil {
		return cfg, err
	}

	cfg.BaseURL = bu.String()
	cfg.AppURL = au.String()

	return cfg, nil
}

//figures out which config to load
func loadConfig(fname string) (*ini.File, error) {
	cwd, _ := os.Getwd()

	cwdPath := fmt.Sprintf(
		"%s/%s",
		cwd,
		".okta",
	)

	hdirPath := fmt.Sprintf(
		"%s/%s",
		os.Getenv("HOME"),
		".okta-aws/config",
	)

	debugCfg("trying to load from config param file")
	if _, err := os.Stat(fname); err == nil {
		debugCfg("loading %s", fname)
		return ini.Load(fname)
	}

	debugCfg("trying to load from CWD")
	if _, err := os.Stat(cwdPath); err == nil {
		debugCfg("loading %s", cwdPath)
		return ini.Load(cwdPath)
	}

	debugCfg("trying to load from home dir")
	if _, err := os.Stat(hdirPath); err == nil {
		debugCfg("loading %s", hdirPath)
		return ini.Load(hdirPath)
	}

	return nil, badCfgErr
}

// loads the aws profile file, which we need
// to look up info to assume roles
func loadAwsCfg() (*ini.File, error) {
	return ini.Load(
		fmt.Sprintf(
			"%s/%s",
			os.Getenv("HOME"),
			".aws/config",
		),
	)
}

// reads your AWS config file to load the role ARN
// for a specific profile; returns the ARN and an error if any
func readAwsProfile(name string) (string, error) {

	var secondRoleArn string
	asec, err := loadAwsCfg()
	if err != nil {
		debugCfg("aws profile load err, %s", err)
		return secondRoleArn, err
	}

	s, err := asec.GetSection(name)
	if err != nil {
		debugCfg("aws profile read err, %s", err)
		return secondRoleArn, err
	}

	if !s.HasKey("role_arn") {
		debugCfg("aws profile %s missing role_arn key", name)
		return secondRoleArn, err
	}

	arnKey, _ := s.GetKey("role_arn")
	secondRoleArn = arnKey.String()

	return secondRoleArn, nil
}