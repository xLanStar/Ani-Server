package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	WebFolder   string `yaml:"webFolder"`
	MediaFolder string `yaml:"mediaFolder"`
	// MediaFile     string `yaml:"mediaFile"`
	UserFolder string `yaml:"userFolder"`
	// UserFile      string `yaml:"userFile"`
	ReviewFolder string `yaml:"reviewFolder"`
	// ReviewFile    string `yaml:"reviewFile"`
	ProfileFolder string `yaml:"profileFolder"`
	DataFile      string `yaml:"dataFile"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func (config *Config) Save(configPath string) {
	data, err := yaml.Marshal(config)

	if err != nil {
		log.Fatal(err)
	}

	err2 := ioutil.WriteFile(configPath, data, 0)

	if err2 != nil {
		log.Fatal(err)
	}
}

func NewDefaultConfig() *Config {
	// Create config structure
	config := &Config{}
	config.DataFile = "./data.bin"
	config.MediaFolder = "./media/"
	config.WebFolder = "./web/"
	config.ProfileFolder = "./profile/"
	config.ReviewFolder = "./review/"
	config.UserFolder = "./user/"
	config.Server.Host = ""
	config.Server.Port = "8050"
	return config
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}
