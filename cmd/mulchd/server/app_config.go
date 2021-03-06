package server

import (
	"errors"
	"fmt"
	"path"

	"github.com/BurntSushi/toml"
)

// AppConfig describes the general configuration of an App
type AppConfig struct {
	// address where the API server will listen
	Listen string

	// URI to libvirtd (qemu only, currently)
	LibVirtURI string

	// translated to a absolute local path (so libvirtd shound run next to us, currently)
	StoragePath string

	// persistent storage (ex: VM database)
	// TODO: create path if needed on startup
	DataPath string

	// prefix for VM names (in libvirt)
	VMPrefix string

	// SSH keys used by Mulch to control & command VMs
	MulchSSHPrivateKey string
	MulchSSHPublicKey  string

	// User (sudoer) created by Mulch in VMs
	MulchSuperUser string

	// Seeds
	Seeds map[string]ConfigSeed

	// global mulchd configuration path
	configPath string
}

// ConfigSeed describes a OS seed
type ConfigSeed struct {
	CurrentURL string
	As         string
}

type tomlAppConfig struct {
	Listen             string
	LibVirtURI         string `toml:"libvirt_uri"`
	StoragePath        string `toml:"storage_path"`
	DataPath           string `toml:"data_path"`
	VMPrefix           string `toml:"vm_prefix"`
	MulchSSHPrivateKey string `toml:"mulch_ssh_private_key"`
	MulchSSHPublicKey  string `toml:"mulch_ssh_public_key"`
	MulchSuperUser     string `toml:"mulch_super_user"`
	Seed               []tomlConfigSeed
}

type tomlConfigSeed struct {
	Name       string
	CurrentURL string `toml:"current_url"`
	As         string
}

// NewAppConfigFromTomlFile return a AppConfig using
// mulchd.toml config file in the given configPath
func NewAppConfigFromTomlFile(configPath string) (*AppConfig, error) {

	filename := path.Clean(configPath + "/mulchd.toml")

	appConfig := &AppConfig{
		configPath: configPath,
		Seeds:      make(map[string]ConfigSeed),
	}

	// defaults (if not in the file)
	tConfig := &tomlAppConfig{
		Listen:         ":8585",
		LibVirtURI:     "qemu:///system",
		StoragePath:    "./var/storage", // example: /srv/mulch
		DataPath:       "./var/data",    // example: /var/lib/mulch
		VMPrefix:       "mulch-",
		MulchSuperUser: "admin",
	}

	if _, err := toml.DecodeFile(filename, tConfig); err != nil {
		return nil, err
	}

	// no check here for most of config elements, it's done later
	appConfig.Listen = tConfig.Listen
	appConfig.LibVirtURI = tConfig.LibVirtURI
	appConfig.StoragePath = tConfig.StoragePath
	appConfig.DataPath = tConfig.DataPath
	appConfig.VMPrefix = tConfig.VMPrefix
	appConfig.MulchSuperUser = tConfig.MulchSuperUser

	if tConfig.MulchSSHPublicKey == "" {
		return nil, errors.New("'mulch_ssh_private_key' config param must be defined")
	}
	appConfig.MulchSSHPrivateKey = tConfig.MulchSSHPrivateKey

	if tConfig.MulchSSHPublicKey == "" {
		return nil, errors.New("'mulch_ssh_public_key' config param must be defined")
	}
	appConfig.MulchSSHPublicKey = tConfig.MulchSSHPublicKey

	for _, seed := range tConfig.Seed {
		if seed.Name == "" {
			return nil, fmt.Errorf("seed 'name' not defined")
		}

		if IsValidTokenName(seed.Name) == false {
			return nil, fmt.Errorf("'%s' is not a valid seed name", seed.Name)
		}

		_, exists := appConfig.Seeds[seed.Name]
		if exists == true {
			return nil, fmt.Errorf("seed name '%s' already defined", seed.Name)
		}

		if seed.CurrentURL == "" {
			return nil, fmt.Errorf("seed '%s': 'current_url' not defined", seed.Name)

		}

		if seed.As == "" {
			return nil, fmt.Errorf("seed '%s': 'as' not defined", seed.Name)
		}

		appConfig.Seeds[seed.Name] = ConfigSeed{
			CurrentURL: seed.CurrentURL,
			As:         seed.As,
		}

	}

	return appConfig, nil
}

// GetTemplateFilepath returns a path to a etc/template file
func (conf *AppConfig) GetTemplateFilepath(name string) string {
	return path.Clean(conf.configPath + "/templates/" + name)
}
