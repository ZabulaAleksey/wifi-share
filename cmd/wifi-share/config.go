package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

type localConfig struct {
	Address  string `json:"address"`
	Root     string `json:"root"`
	Data     string `json:"data"`
	Web      string `json:"web"`
	Password string `json:"password"`
}

func loadLocalConfig(path string) (localConfig, error) {
	config := localConfig{
		Address: ":8080",
		Root:    "./shared",
		Data:    "./data",
		Web:     "./web/dist",
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, fmt.Errorf("%s is missing; copy config.example.json and set a password", path)
		}
		return config, err
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return config, fmt.Errorf("read %s: %w", path, err)
	}
	if strings.TrimSpace(config.Password) == "" {
		return config, fmt.Errorf("password is empty in %s", path)
	}
	return config, nil
}

func accessURLs(address string) []string {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil
	}
	result := []string{"http://localhost:" + port}
	interfaces, err := net.Interfaces()
	if err != nil {
		return result
	}
	for _, networkInterface := range interfaces {
		if networkInterface.Flags&net.FlagUp == 0 || networkInterface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, err := networkInterface.Addrs()
		if err != nil {
			continue
		}
		for _, item := range addresses {
			parsed, err := url.Parse("http://" + item.String())
			if err != nil {
				continue
			}
			ip := net.ParseIP(parsed.Hostname())
			if ip == nil || ip.To4() == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			result = append(result, "http://"+net.JoinHostPort(ip.String(), port))
		}
	}
	return result
}
