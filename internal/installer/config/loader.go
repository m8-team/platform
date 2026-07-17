package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	"sigs.k8s.io/yaml"
)

type InstallationFile struct {
	Path         string
	Installation installerv1alpha1.PlatformInstallation
	Digest       string
}

func LoadInstallationFile(path string) (InstallationFile, error) {
	if path == "" {
		return InstallationFile{}, fmt.Errorf("installation file path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return InstallationFile{}, fmt.Errorf("read installation file %q: %w", path, err)
	}

	var installation installerv1alpha1.PlatformInstallation
	if err := yaml.Unmarshal(data, &installation); err != nil {
		return InstallationFile{}, fmt.Errorf("parse installation file %q: %w", path, err)
	}

	installation = installation.Defaulted()
	if err := installation.Validate(); err != nil {
		return InstallationFile{}, err
	}

	digest, err := DigestObject(installation)
	if err != nil {
		return InstallationFile{}, err
	}

	return InstallationFile{
		Path:         path,
		Installation: installation,
		Digest:       digest,
	}, nil
}

func DigestObject(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("canonicalize object: %w", err)
	}

	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
