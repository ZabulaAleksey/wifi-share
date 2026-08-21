package main

import (
	"encoding/json"
	"fmt"
	"net"
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
		Address: "127.0.0.1:8080",
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

func saveLocalConfig(path string, config localConfig) error {
	raw, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, byte(10))
	return os.WriteFile(path, raw, 0o600)
}

func accessURLs(address string) []string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil
	}
	if host == "0.0.0.0" {
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
			for _, address := range addresses {
				ip, _, err := net.ParseCIDR(address.String())
				if err != nil || ip == nil || !ip.IsPrivate() {
					continue
				}
				result = append(result, "http://"+net.JoinHostPort(ip.String(), port))
			}
		}
		return result
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		host = "localhost"
	}
	return []string{"http://" + net.JoinHostPort(host, port)}
}
