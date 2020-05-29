package pkg

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
)

const (
	testUser  = "Test User"
	testEmail = "testuser@example.com"
)

func newRepoEmpty(t *testing.T) *Repo {
	dir := newTempDir(t)

	repo, err := git.PlainInit(dir, false)
	checkFatal(t, err)

	return newRepo(repo, dir)
}

func newRepoWithUntracked(t *testing.T) *Repo {
	r := newRepoEmpty(t)
	r.writeFile(t, "README", "I'm a README file")

	return r
}

func newRepoWithStaged(t *testing.T) *Repo {
	r := newRepoEmpty(t)
	r.writeFile(t, "README", "I'm a README file")
	r.addFile(t, "README")

	return r
}

func newRepoWithCommit(t *testing.T) *Repo {
	r := newRepoEmpty(t)
	r.writeFile(t, "README", "I'm a README file")
	r.addFile(t, "README")
	r.newCommit(t, "Initial commit")

	return r
}

func newRepoWithModified(t *testing.T) *Repo {
	r := newRepoEmpty(t)
	r.writeFile(t, "README", "I'm a README file")
	r.addFile(t, "README")
	r.newCommit(t, "Initial commit")
	r.writeFile(t, "README", "I'm modified")

	return r
}

func newRepoWithIgnored(t *testing.T) *Repo {
	r := newRepoEmpty(t)
	r.writeFile(t, ".gitignore", "ignoreme")
	r.addFile(t, ".gitignore")
	r.newCommit(t, "Initial commit")
	r.writeFile(t, "ignoreme", "I'm being ignored")

	return r
}

func newRepoWithLocalBranch(t *testing.T) *Repo {
	r := newRepoWithCommit(t)
	r.newBranch(t, "local")
	return r
}

func newRepoWithClonedBranch(t *testing.T) *Repo {
	origin := newRepoWithCommit(t)

	r := origin.clone(t)
	r.newBranch(t, "local")

	return r
}

func newRepoWithBranchAhead(t *testing.T) *Repo {
	origin := newRepoWithCommit(t)

	r := origin.clone(t)
	r.writeFile(t, "new", "I'm a new file")
	r.addFile(t, "new")
	r.newCommit(t, "new commit")

	return r
}

func newRepoWithBranchBehind(t *testing.T) *Repo {
	origin := newRepoWithCommit(t)

	r := origin.clone(t)

	origin.writeFile(t, "origin.new", "I'm a new file on origin")
	origin.addFile(t, "origin.new")
	origin.newCommit(t, "new origin commit")

	r.fetch(t)
	return r
}

func newRepoWithBranchAheadAndBehind(t *testing.T) *Repo {
	origin := newRepoWithCommit(t)

	r := origin.clone(t)
	r.writeFile(t, "local.new", "I'm a new file on local")
	r.addFile(t, "local.new")
	r.newCommit(t, "new local commit")

	origin.writeFile(t, "origin.new", "I'm a new file on origin")
	origin.addFile(t, "origin.new")
	origin.newCommit(t, "new origin commit")

	r.fetch(t)
	return r
}

func newTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "git-get-repo-")
	checkFatal(t, errors.Wrap(err, "Failed creating test repo directory"))

	// Automatically remove repo when test is over
	t.Cleanup(func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Errorf("failed cleaning up repo")
		}
	})

	return dir
}

func (r *Repo) writeFile(t *testing.T, name string, content string) {
	wt, err := r.repo.Worktree()
	checkFatal(t, errors.Wrap(err, "Failed getting workree"))

	file, err := wt.Filesystem.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	checkFatal(t, errors.Wrap(err, "Failed opening a file"))

	_, err = file.Write([]byte(content))
	checkFatal(t, errors.Wrap(err, "Failed writing a file"))
}

func (r *Repo) addFile(t *testing.T, name string) {
	wt, err := r.repo.Worktree()
	checkFatal(t, errors.Wrap(err, "Failed getting workree"))

	_, err = wt.Add(name)
	checkFatal(t, errors.Wrap(err, "Failed adding file to index"))
}

func (r *Repo) newCommit(t *testing.T, msg string) {
	wt, err := r.repo.Worktree()
	checkFatal(t, errors.Wrap(err, "Failed getting workree"))

	opts := &git.CommitOptions{
		Author: &object.Signature{
			Name:  testUser,
			Email: testEmail,
			When:  time.Date(2000, 01, 01, 16, 00, 00, 0, time.UTC),
		},
	}

	_, err = wt.Commit(msg, opts)
	checkFatal(t, errors.Wrap(err, "Failed creating commit"))
}

func (r *Repo) newBranch(t *testing.T, name string) {
	head, err := r.repo.Head()
	checkFatal(t, err)

	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(name), head.Hash())

	err = r.repo.Storer.SetReference(ref)
	checkFatal(t, err)
}

func (r *Repo) clone(t *testing.T) *Repo {
	dir := newTempDir(t)
	url, err := ParseURL("file://" + r.path)
	checkFatal(t, err)

	repo, err := CloneRepo(url, dir, true)
	checkFatal(t, err)

	return repo
}

func (r *Repo) fetch(t *testing.T) {
	err := r.Fetch()
	checkFatal(t, err)
}

func checkFatal(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("%+v", err)
	}
}
