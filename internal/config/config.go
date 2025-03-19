package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type Cfg struct {
	MetricsPrefix        string
	GraphqlURL           string
	GraphqlAPIToken      string
	GraphqlCustomHeaders []CustomHeader
	CacheExpire          int64
	QueryTimeout         int64
	FailFast             bool
	ExtendCacheOnError   bool
	Queries              []Query
}

type Query struct {
	Query   string
	Metrics []Metric
}

type Metric struct {
	Description string
	Placeholder string
	Labels      []string
	Value       string
	Name        string
}

type CustomHeader struct {
	Key   string
	Value string
}

var (
	Config     *Cfg
	ConfigPath string
)

func Init(configPath string) error {
	ConfigPath = configPath
	content := []byte(`{}`)
	_, err := os.Stat(ConfigPath)
	if !os.IsNotExist(err) {
		content, err = os.ReadFile(ConfigPath)
		if err != nil {
			return err
		}
	}

	if len(content) == 0 {
		content = []byte(`{}`)
	}

	err = json.Unmarshal(content, &Config)
	if err != nil {
		return err
	}

	tokenVal, tokenIsSet := os.LookupEnv("GRAPHQLAPITOKEN")

	if Config.GraphqlCustomHeaders == nil {
		Config.GraphqlCustomHeaders = []CustomHeader{
			{
				Key:   "Content-Type",
				Value: "application/x-www-form-urlencoded",
			},
		}
		if tokenIsSet {
			Config.GraphqlCustomHeaders = append(
				Config.GraphqlCustomHeaders,
				CustomHeader{
					Key:   "Authorization",
					Value: "%s",
				},
			)
		}
	}

	for i := range Config.GraphqlCustomHeaders {
		if Config.GraphqlCustomHeaders[i].Key == "Authorization" {
			if tokenIsSet {
				Config.GraphqlCustomHeaders[i].Value = fmt.Sprintf(Config.GraphqlCustomHeaders[i].Value, tokenVal)
			}
		}
	}

	if Config.QueryTimeout == 0 {
		Config.QueryTimeout = 60
	}

	slog.Info(fmt.Sprintf("Finished reading config from %s", configPath))
	return nil
}
