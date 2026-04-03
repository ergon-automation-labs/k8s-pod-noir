package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Cluster struct {
	Context   string `yaml:"context"`
	Namespace string `yaml:"namespace"`
}

type Detective struct {
	Name string `yaml:"name"`
}

type Config struct {
	Cluster   Cluster   `yaml:"cluster"`
	Detective Detective `yaml:"detective"`
}

func Default() Config {
	return Config{
		Cluster: Cluster{
			Context:   "",
			Namespace: "pod-noir",
		},
		Detective: Detective{
			Name: "Sam Reyes",
		},
	}
}

func Load(path string) (Config, error) {
	c := Default()
	if path == "" {
		return c, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return c, nil
		}
		return c, err
	}
	if err := yaml.Unmarshal(b, &c); err != nil {
		return c, fmt.Errorf("yaml: %w", err)
	}
	if c.Cluster.Namespace == "" {
		c.Cluster.Namespace = "pod-noir"
	}
	return c, nil
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".pod-noir", "config.yaml")
}
