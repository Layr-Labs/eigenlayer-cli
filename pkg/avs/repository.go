package avs

import (
	"crypto/sha1"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/urfave/cli/v2"
)

const (
	RemoteSpecificationRepositoryURL = "https://raw.githubusercontent.com/Layr-Labs/eigenlayer-cli/refs/heads/feat/slashing/pkg/avs/specs"
)

var (
	//go:embed specs/*
	EmbeddedRepository  embed.FS
	RepositorySubFolder = ".eigenlayer/avs/specs"
)

type Repository struct {
	initialized bool
	logger      logging.Logger
	path        string
}

func NewRepository(ctx *cli.Context) (*Repository, error) {
	repo := Repository{}
	repo.initialized = false
	repo.logger = common.GetLogger(ctx)

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	repo.path = filepath.Clean(filepath.Join(home, RepositorySubFolder))

	return &repo, nil
}

func (repo *Repository) init() error {
	if repo.initialized {
		return nil
	}

	repo.logger.Debug(fmt.Sprintf("Initializing local repository [Path=%s]", repo.path))
	if err := os.MkdirAll(repo.path, os.ModePerm); err != nil {
		return err
	}

	files, err := os.ReadDir(repo.path)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		if err := repo.extract("specs"); err != nil {
			return err
		}
	}

	return nil
}

func (repo *Repository) extract(path string) error {
	dir := filepath.Clean(filepath.Join(repo.path, "..", path))
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil
	}

	entries, err := EmbeddedRepository.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		origin := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if err := repo.extract(origin); err != nil {
				return err
			}
		} else {
			content, err := EmbeddedRepository.ReadFile(origin)
			if err != nil {
				return err
			}

			repo.logger.Debug(fmt.Sprintf("Extracting: %s", origin))
			if err = os.WriteFile(filepath.Clean(filepath.Join(dir, entry.Name())), content, os.ModePerm); err != nil {
				return err
			}
		}
	}

	return nil
}

func (repo *Repository) List() (*[]*adapters.BaseSpecification, error) {
	if err := repo.init(); err != nil {
		return nil, err
	}

	repo.logger.Debug(fmt.Sprintf("Listing specifications in local repository [Path=%s]", repo.path))

	var specifications []*adapters.BaseSpecification
	err := filepath.WalkDir(repo.path, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !file.IsDir() && file.Name() == "avs.json" {
			file, err := os.Open(path)
			if err != nil {
				return err
			}

			data, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			spec, err := adapters.NewBaseSpecification(data)
			if err != nil {
				return err
			}

			name := filepath.Dir(path)[len(repo.path)+1:]
			if name != spec.Name {
				return fmt.Errorf("specification name does not match repository location: %s", name)
			}

			specifications = append(specifications, spec)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &specifications, nil
}

func (repo *Repository) Reset() error {
	repo.logger.Debug(fmt.Sprintf("Resetting local specification repository [Path=%s]", repo.path))

	if err := os.RemoveAll(repo.path); err != nil {
		return err
	}

	repo.initialized = false
	return repo.init()
}

func (repo *Repository) Update() error {
	repo.logger.Debug(
		fmt.Sprintf(
			"Updating local specification repository [Path=%s, RemoteRepository=%s]",
			repo.path,
			RemoteSpecificationRepositoryURL,
		),
	)

	manifestUrl := RemoteSpecificationRepositoryURL + "/manifest"
	repo.logger.Info(fmt.Sprintf("Downloading: %s", manifestUrl))
	manifest, err := download(manifestUrl)
	if err != nil {
		return err
	}

	entries := strings.Split(string(manifest), "\n")
	for _, entry := range entries {
		if entry != "" {
			tokens := strings.FieldsFunc(entry, func(r rune) bool { return r == ' ' })
			hash := strings.TrimSpace(tokens[0])
			entry := strings.TrimSpace(tokens[1])

			path := filepath.Clean(filepath.Join(repo.path, entry))
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return nil
			}

			file, err := os.Open(path)
			if err == nil {
				data, err := io.ReadAll(file)
				if err == nil {
					if strings.EqualFold(hash, fmt.Sprintf("%x", sha1.Sum(data))) {
						repo.logger.Info(fmt.Sprintf("Skipping (no changes): %s", entry))
						continue
					}
				}
			}

			repo.logger.Info(fmt.Sprintf("Downloading: %s", entry))
			content, err := download(RemoteSpecificationRepositoryURL + "/" + entry)
			if err != nil {
				return err
			}

			if err = os.WriteFile(path, content, os.ModePerm); err != nil {
				return err
			}
		}
	}

	return nil
}

func (repo *Repository) LoadResource(name string, resource string) ([]byte, error) {
	path := filepath.Clean(filepath.Join(repo.path, filepath.Join(name, resource)))
	repo.logger.Debug(fmt.Sprintf("Loading specification resource [Path=%s]", path))

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(file)
}

func (repo *Repository) LoadPlugin(name string, url string) (*plugin.Plugin, error) {
	path := filepath.Clean(filepath.Join(repo.path, filepath.Join(name, "plugin.so")))
	repo.logger.Debug(fmt.Sprintf("Loading specification plugin [Path=%s]", path))

	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		repo.logger.Debug(fmt.Sprintf("Plugin not found [Path=%s]", path))

		repo.logger.Debug(fmt.Sprintf("Downloading plugin [URL=%s]", url))
		data, err := download(url)
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(path, data, os.ModePerm); err != nil {
			return nil, err
		}

		repo.logger.Debug(fmt.Sprintf("Plugin downloaded [Path=%s, Size=%d]", path, len(data)))
	}

	return plugin.Open(path)
}

func download(url string) ([]byte, error) {
	client := http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
		redirect, err := resp.Location()
		if err != nil {
			return []byte{}, err
		}

		return download(redirect.String())
	} else if resp.StatusCode >= 400 {
		return []byte{}, fmt.Errorf("error fetching url: %s", resp.Status)
	}

	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			fmt.Println("error closing url body")
		}
	}(resp.Body)

	response, err := io.ReadAll(http.MaxBytesReader(nil, resp.Body, 10000000))
	if err != nil {
		return nil, err
	}

	return response, nil
}
