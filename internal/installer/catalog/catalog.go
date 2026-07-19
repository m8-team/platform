package catalog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	installerv1alpha1 "github.com/m8-team/platform/api/installer/v1alpha1"
	"github.com/m8-team/platform/internal/installer/config"
	"sigs.k8s.io/yaml"
)

var (
	ErrReleaseNotFound = errors.New("platform release not found")
	ErrUnsignedRelease = errors.New("platform release is not signed")
)

type ReleaseCatalog interface {
	Resolve(ctx context.Context, version string) (installerv1alpha1.PlatformRelease, error)
	Verify(ctx context.Context, release installerv1alpha1.PlatformRelease) error
	Digest(ctx context.Context, release installerv1alpha1.PlatformRelease) (string, error)
}

type FileCatalog struct {
	Dir           string
	AllowUnsigned bool
}

func NewFileCatalog(dir string) FileCatalog {
	if dir == "" {
		dir = "catalog/releases"
	}
	return FileCatalog{Dir: dir}
}

func (c FileCatalog) Resolve(ctx context.Context, version string) (installerv1alpha1.PlatformRelease, error) {
	if err := ctx.Err(); err != nil {
		return installerv1alpha1.PlatformRelease{}, err
	}
	if version == "" {
		return installerv1alpha1.PlatformRelease{}, fmt.Errorf("release version is required")
	}

	path := filepath.Join(c.Dir, version+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return installerv1alpha1.PlatformRelease{}, fmt.Errorf("%w: %s", ErrReleaseNotFound, version)
		}
		return installerv1alpha1.PlatformRelease{}, fmt.Errorf("read release catalog %q: %w", path, err)
	}

	var release installerv1alpha1.PlatformRelease
	if err := yaml.Unmarshal(data, &release); err != nil {
		return installerv1alpha1.PlatformRelease{}, fmt.Errorf("parse release catalog %q: %w", path, err)
	}
	if release.APIVersion == "" {
		release.APIVersion = installerv1alpha1.GroupName + "/" + installerv1alpha1.Version
	}
	if release.Kind == "" {
		release.Kind = "PlatformRelease"
	}
	if release.Name == "" {
		release.Name = version
	}
	if err := release.Validate(); err != nil {
		return installerv1alpha1.PlatformRelease{}, err
	}

	return release, nil
}

func (c FileCatalog) Verify(ctx context.Context, release installerv1alpha1.PlatformRelease) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := release.Validate(); err != nil {
		return err
	}
	if !c.AllowUnsigned && release.Spec.Signature.Value == "" {
		return ErrUnsignedRelease
	}
	return nil
}

func (c FileCatalog) Digest(ctx context.Context, release installerv1alpha1.PlatformRelease) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return config.DigestObject(release)
}
