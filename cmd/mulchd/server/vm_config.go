package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/c2h5oh/datasize"
)

// VMConfig stores needed parameters for a new VM
type VMConfig struct {
	FileContent string // config file content

	Name           string
	Hostname       string
	Timezone       string
	AppUser        string
	Seed           string
	InitUpgrade    bool
	DiskSize       uint64
	RAMSize        uint64
	CPUCount       int
	Env            map[string]string
	BackupDiskSize uint64
	RestoreBackup  string

	Prepare []*VMConfigScript
	Backup  []*VMConfigScript
	Restore []*VMConfigScript
	// + save scripts
	// + restore scripts
}

// VMConfigScript is a script for prepare, save and restore steps
type VMConfigScript struct {
	ScriptURL string
	As        string
}

type tomlVMConfig struct {
	Name           string
	Hostname       string
	Timezone       string
	AppUser        string `toml:"app_user"`
	Seed           string
	InitUpgrade    bool              `toml:"init_upgrade"`
	DiskSize       datasize.ByteSize `toml:"disk_size"`
	RAMSize        datasize.ByteSize `toml:"ram_size"`
	CPUCount       int               `toml:"cpu_count"`
	Env            [][]string
	BackupDiskSize datasize.ByteSize `toml:"backup_disk_size"`
	RestoreBackup  string            `toml:"restore_backup"`

	PreparePrefixURL string `toml:"prepare_prefix_url"`
	Prepare          []string
	BackupPrefixURL  string `toml:"backup_prefix_url"`
	Backup           []string
	RestorePrefixURL string `toml:"restore_prefix_url"`
	Restore          []string
}

func vmConfigGetScript(tScript string, prefixURL string) (*VMConfigScript, error) {
	script := &VMConfigScript{}

	sepPlace := strings.Index(tScript, "@")
	if sepPlace == -1 {
		return nil, fmt.Errorf("prepre line should use the 'user@url' format ('%s')", tScript)
	}

	as := tScript[:sepPlace]
	scriptURL := prefixURL + tScript[sepPlace+1:]

	if !IsValidTokenName(as) {
		return nil, fmt.Errorf("'%s' is not a valid user name", as)
	}
	script.As = as

	// test readability
	stream, errG := GetScriptFromURL(scriptURL)
	if errG != nil {
		return nil, fmt.Errorf("unable to get script '%s': %s", scriptURL, errG)
	}
	defer stream.Close()

	// check script signature
	signature := make([]byte, 2)
	n, errR := stream.Read(signature)
	if n != 2 || errR != nil {
		return nil, fmt.Errorf("error reading script '%s' (n=%d)", scriptURL, n)
	}
	if string(signature) != "#!" {
		return nil, fmt.Errorf("script '%s': no shebang found, is it really a shell script?", scriptURL)
	}

	script.ScriptURL = scriptURL
	return script, nil
}

// NewVMConfigFromTomlReader cretes a new VMConfig instance from
// a io.Reader containing VM configuration description
func NewVMConfigFromTomlReader(configIn io.Reader) (*VMConfig, error) {
	content, err := ioutil.ReadAll(configIn)
	if err != nil {
		return nil, err
	}

	vmConfig := &VMConfig{
		Env:         make(map[string]string),
		FileContent: string(content),
	}

	// defaults (if not in the file)
	tConfig := &tomlVMConfig{
		Hostname:       "localhost.localdomain",
		Timezone:       "Europe/Paris",
		AppUser:        "app",
		InitUpgrade:    true,
		CPUCount:       1,
		BackupDiskSize: 2 * datasize.GB,
	}

	if _, err := toml.Decode(vmConfig.FileContent, tConfig); err != nil {
		return nil, err
	}

	if tConfig.Name == "" || !IsValidTokenName(tConfig.Name) {
		return nil, fmt.Errorf("invalid VM name '%s'", tConfig.Name)
	}
	vmConfig.Name = tConfig.Name

	vmConfig.Hostname = tConfig.Hostname
	vmConfig.Timezone = tConfig.Timezone

	if tConfig.AppUser == "" {
		return nil, fmt.Errorf("invalid app_user name '%s'", tConfig.AppUser)
	}
	vmConfig.AppUser = tConfig.AppUser

	if tConfig.Seed == "" || !IsValidTokenName(tConfig.Seed) {
		return nil, fmt.Errorf("invalid seed image '%s'", tConfig.Seed)
	}
	vmConfig.Seed = tConfig.Seed

	vmConfig.InitUpgrade = tConfig.InitUpgrade

	if tConfig.DiskSize < 1*datasize.MB {
		return nil, fmt.Errorf("looks like a too small disk (%s)", tConfig.DiskSize)
	}
	vmConfig.DiskSize = tConfig.DiskSize.Bytes()

	if tConfig.RAMSize < 1*datasize.MB {
		return nil, fmt.Errorf("looks like a too small RAM amount (%s)", tConfig.RAMSize)
	}
	vmConfig.RAMSize = tConfig.RAMSize.Bytes()

	if tConfig.CPUCount < 1 {
		return nil, fmt.Errorf("need a least one CPU")
	}
	vmConfig.CPUCount = tConfig.CPUCount

	for _, line := range tConfig.Env {
		if len(line) != 2 {
			return nil, fmt.Errorf("invalid 'env' line, need two values (key, val), found %d", len(line))
		}

		key := line[0]
		val := line[1]
		if !IsValidTokenName(key) {
			return nil, fmt.Errorf("invalid 'env' name '%s'", key)
		}

		// TODO: check for reserved names?

		_, exists := vmConfig.Env[key]
		if exists == true {
			return nil, fmt.Errorf("duplicated 'env' name '%s'", key)
		}

		vmConfig.Env[key] = val
	}

	if tConfig.BackupDiskSize < 32*datasize.MB {
		return nil, fmt.Errorf("looks like a too small backup disk (%s)", tConfig.BackupDiskSize)
	}
	vmConfig.BackupDiskSize = tConfig.BackupDiskSize.Bytes()

	for _, tScript := range tConfig.Prepare {
		script, err := vmConfigGetScript(tScript, tConfig.PreparePrefixURL)
		if err != nil {
			return nil, err
		}
		vmConfig.Prepare = append(vmConfig.Prepare, script)
	}

	for _, tScript := range tConfig.Backup {
		script, err := vmConfigGetScript(tScript, tConfig.BackupPrefixURL)
		if err != nil {
			return nil, err
		}
		vmConfig.Backup = append(vmConfig.Backup, script)
	}

	for _, tScript := range tConfig.Restore {
		script, err := vmConfigGetScript(tScript, tConfig.RestorePrefixURL)
		if err != nil {
			return nil, err
		}
		vmConfig.Restore = append(vmConfig.Restore, script)
	}
	vmConfig.RestoreBackup = tConfig.RestoreBackup

	return vmConfig, nil
}
