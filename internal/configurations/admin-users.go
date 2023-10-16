package configurations

import (
	"fmt"
	"github.com/json-iterator/go" // JSONC library
	"gopkg.in/yaml.v2"            // YAML library
	"os"
	"path/filepath"
	"strings"
)

type (
	Config struct {
		AdminUsers StringList       `json:"admin-users" yaml:"admin_users"`
		UsersLimit UserLimitMap     `json:"user-limit" yaml:"user_limit"`
		PamConfig  ConfigurationMap `json:"pam-config,omitempty" yaml:"pam_config,omitempty"`
	}
	StringList       []string
	UserLimitMap     map[string]int
	ConfigurationMap map[string]interface{}
)

var (
	AdminUsers StringList = []string{
		"root",
		"sadeq",
	}

	UsersLimit = UserLimitMap{
		"guest":   1,
		"anis":    1,
		"farnaz":  2,
		"sepidar": 3,
	}

	PamConfig = ConfigurationMap{
		"debug":     true,
		"pseudo":    true,
		"report-to": "",
	}
)

func LoadConfigFile(filename string, config *Config) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Check the file extension to determine the format (JSONC or YAML)
	switch ext := filepath.Ext(filename); ext {
	case ".jsonc", ".json":
		err = jsoniter.Unmarshal(data, &config)
		if err != nil {
			return err
		}
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	return nil
}

func (l StringList) IsInList(s string) bool {
	for _, name := range l {
		if strings.ToLower(s) == strings.ToLower(name) {
			return true
		}
	}
	return false
}

func (m UserLimitMap) GetUserLimits(name string) int {
	limit, find := m[name]
	if !find {
		return 0
	}
	return limit
}
