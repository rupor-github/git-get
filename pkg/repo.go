package pkg

import (
	"github.com/pkg/errors"

	git "github.com/libgit2/git2go/v30"
)

type Repo struct {
	repo   *git.Repository
	Status *RepoStatus
}

func CloneRepo(url string, path string) error {
	options := &git.CloneOptions{
		CheckoutOpts:         nil,
		FetchOptions:         nil,
		Bare:                 false,
		CheckoutBranch:       "",
		RemoteCreateCallback: nil,
	}

	_, err := git.Clone(url, path, options)
	if err != nil {
		return errors.Wrap(err, "Failed cloning repo")
	}
	return nil
}

func OpenRepo(path string) (*Repo, error) {
	r, err := git.OpenRepository(path)
	if err != nil {
		return nil, errors.Wrap(err, "Failed opening repo")
	}

	repoStatus, err := loadStatus(r)
	if err != nil {
		return nil, err
	}

	repo := &Repo{
		repo:   r,
		Status: repoStatus,
	}

	return repo, nil
}

func (r *Repo) Reload() error {
	status, err := loadStatus(r.repo)
	if err != nil {
		return err
	}

	r.Status = status
	return nil
}

func (r *Repo) Fetch() error {
	remoteNames, err := r.repo.Remotes.List()
	if err != nil {
		return errors.Wrap(err, "Failed listing remoteNames")
	}

	for _, name := range remoteNames {
		remote, err := r.repo.Remotes.Lookup(name)
		if err != nil {
			return errors.Wrap(err, "Failed looking up remote")
		}

		err = remote.Fetch(nil, nil, "")
		if err != nil {
			return errors.Wrap(err, "Failed fetching remote")
		}
	}

	return nil
}
