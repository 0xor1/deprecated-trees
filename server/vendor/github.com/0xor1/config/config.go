package config

import (
	"github.com/0xor1/json"
	"os"
	"strings"
	"time"
)

type config struct {
	defaults               *json.Json
	fileValues             *json.Json
	envVarsStringSeparator string
}

//create a new config object based on a env vars and/or a json file and/or programmatic defaults
//pass in empty file path to not use a config file, pass in an empty envVarSeparator to ignore environment variables
func New(file string, envVarSeparator string) *config {
	ret := &config{
		defaults:               json.MustNew(),
		fileValues:             json.MustNew(),
		envVarsStringSeparator: envVarSeparator,
	}
	if file != "" {
		ret.fileValues = json.MustFromFile(file)
	}
	return ret
}

func (c *config) SetDefault(path string, val interface{}) {
	jsonPath, _, _ := makeJsonPathAndGetEnvValAndExists(path, "")
	c.defaults.Set(val, jsonPath...)
}

func (c *config) GetString(path string) string {
	if jsonPath, envVal, envValExists := makeJsonPathAndGetEnvValAndExists(path, c.envVarsStringSeparator); envValExists {
		return envVal
	} else {
		if val, err := c.fileValues.String(jsonPath...); err == nil {
			return val
		}
		return c.defaults.MustString(jsonPath...)
	}
}

func (c *config) GetStringSlice(path string) []string {
	if jsonPath, envVal, envValExists := makeJsonPathAndGetEnvValAndExists(path, c.envVarsStringSeparator); envValExists {
		return json.MustFromString(envVal).MustStringSlice()
	} else {
		if val, err := c.fileValues.StringSlice(jsonPath...); err == nil {
			return val
		}
		return c.defaults.MustStringSlice(jsonPath...)
	}
}

func (c *config) GetMap(path string) map[string]interface{} {
	if jsonPath, envVal, envValExists := makeJsonPathAndGetEnvValAndExists(path, c.envVarsStringSeparator); envValExists {
		return json.MustFromString(envVal).MustMap()
	} else {
		if val, err := c.fileValues.Map(jsonPath...); err == nil {
			return val
		}
		return c.defaults.MustMap(jsonPath...)
	}
}

func (c *config) GetStringMap(path string) map[string]string {
	if jsonPath, envVal, envValExists := makeJsonPathAndGetEnvValAndExists(path, c.envVarsStringSeparator); envValExists {
		return json.MustFromString(envVal).MustMapString()
	} else {
		if val, err := c.fileValues.MapString(jsonPath...); err == nil {
			return val
		}
		return c.defaults.MustMapString(jsonPath...)
	}
}

func (c *config) GetInt(path string) int {
	if jsonPath, envVal, envValExists := makeJsonPathAndGetEnvValAndExists(path, c.envVarsStringSeparator); envValExists {
		return json.MustFromString(envVal).MustInt()
	} else {
		if val, err := c.fileValues.Int(jsonPath...); err == nil {
			return val
		}
		return c.defaults.MustInt(jsonPath...)
	}
}

func (c *config) GetInt64(path string) int64 {
	if jsonPath, envVal, envValExists := makeJsonPathAndGetEnvValAndExists(path, c.envVarsStringSeparator); envValExists {
		return json.MustFromString(envVal).MustInt64()
	} else {
		if val, err := c.fileValues.Int64(jsonPath...); err == nil {
			return val
		}
		return c.defaults.MustInt64(jsonPath...)
	}
}

func (c *config) GetBool(path string) bool {
	if jsonPath, envVal, envValExists := makeJsonPathAndGetEnvValAndExists(path, c.envVarsStringSeparator); envValExists {
		return json.MustFromString(envVal).MustBool()
	} else {
		if val, err := c.fileValues.Bool(jsonPath...); err == nil {
			return val
		}
		return c.defaults.MustBool(jsonPath...)
	}
}

func (c *config) GetTime(path string) time.Time {
	if jsonPath, envVal, envValExists := makeJsonPathAndGetEnvValAndExists(path, c.envVarsStringSeparator); envValExists {
		return json.MustFromString(envVal).MustTime()
	} else {
		if val, err := c.fileValues.Time(jsonPath...); err == nil {
			return val
		}
		return c.defaults.MustTime(jsonPath...)
	}
}

func (c *config) GetDuration(path string) time.Duration {
	if jsonPath, envVal, envValExists := makeJsonPathAndGetEnvValAndExists(path, c.envVarsStringSeparator); envValExists {
		return json.MustFromString(envVal).MustDuration()
	} else {
		if val, err := c.fileValues.Duration(jsonPath...); err == nil {
			return val
		}
		return c.defaults.MustDuration(jsonPath...)
	}
}

func makeJsonPathAndGetEnvValAndExists(path, envVarsStringSeparator string) ([]interface{}, string, bool) {
	parts := strings.Split(path, ".")
	envName := ""
	if envVarsStringSeparator != "" {
		envName = strings.ToUpper(strings.Join(parts, envVarsStringSeparator))
	}
	if envName != "" {
		if val, exists := os.LookupEnv(envName); exists {
			return nil, val, true
		}
	}

	is := make([]interface{}, 0, len(parts))
	for _, str := range parts {
		is = append(is, str)
	}
	return is, "", false
}
