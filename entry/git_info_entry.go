package rkentry

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-common/common"
	"os"
	"path"
	"strings"
)

const (
	GitInfoEntryName        = "GitInfoDefault"
	GitInfoEntryType        = "GitInfoEntry"
	GitInfoEntryDescription = "Internal RK entry which describes git information."
)

// Bootstrap config of application's basic information.

type BootConfigGitInfo struct {
	Package string `yaml:"package" json:"package"`
	Url     string `yaml:"url" json:"url"`
	Branch  string `yaml:"branch" json:"branch"`
	Tag     string `yaml:"tag" json:"tag"`
	Commit  struct {
		ID        string `yaml:"id" json:"id"`
		Date      string `yaml:"date" json:"date"`
		Abbr      string `yaml:"abbr" json:"abbr"`
		Sub       string `yaml:"sub" json:"sub"`
		Committer struct {
			Name  string `yaml:"name" json:"name"`
			Email string `yaml:"email" json:"email"`
		} `yaml:"committer" json:"committer"`
	} `yaml:"commit" json:"commit"`
}

// AppInfo Entry contains bellow fields.
type GitInfoEntry struct {
	EntryName        string `json:"entryName" yaml:"entryName"`
	EntryType        string `json:"entryType" yaml:"entryType"`
	EntryDescription string `json:"entryDescription" yaml:"entryDescription"`
	Package          string `json:"package" yaml:"package"`
	Url              string `json:"url" yaml:"url"`
	Branch           string `yaml:"branch" json:"branch"`
	Tag              string `yaml:"tag" json:"tag"`
	CommitId         string `yaml:"commitId" json:"commitId"`
	CommitIdAbbr     string `yaml:"commitIdAbbr" json:"commitIdAbbr"`
	CommitDate       string `yaml:"commitDate" json:"commitDate"`
	CommitSub        string `yaml:"commitSub" json:"commitSub"`
	CommitterName    string `yaml:"committerName" json:"committerName"`
	CommitterEmail   string `yaml:"committerEmail" json:"committerEmail"`
}

// GitInfo Entry Option which used while registering entry from codes.
type GitInfoEntryOption func(*GitInfoEntry)

// Provide package name.
func WithPackageGitInfo(name string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.Package = name
	}
}

// Provide git url.
func WithUrlGitInfo(url string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.Url = url
	}
}

// Provide branch.
func WithBranchGitInfo(branch string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.Branch = branch
	}
}

// Provide tag.
func WithTagGitInfo(tag string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.Tag = tag
	}
}

// Provide commit id.
func WithCommitIdGitInfo(id string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.CommitId = id
	}
}

// Provide commit id abbr.
func WithCommitIdAbbrGitInfo(id string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.CommitIdAbbr = id
	}
}

// Provide date.
func WithCommitDateInfo(date string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.CommitDate = date
	}
}

// Provide subject.
func WithCommitSubGitInfo(sub string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.CommitSub = sub
	}
}

// Provide committer name.
func WithCommitterNameGitInfo(name string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.CommitterName = name
	}
}

// Provide committer email.
func WithCommitterEmailGitInfo(email string) GitInfoEntryOption {
	return func(entry *GitInfoEntry) {
		entry.CommitterEmail = email
	}
}

// Implements rkentry.EntryRegFunc which generate RKEntry based on boot configuration file.
func RegisterGitInfoEntriesFromConfig(string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: Unmarshal user provided config into boot config struct
	config := &BootConfigGitInfo{}

	// We will looking for git.yaml file in current working directory
	//
	// In case, there are no git.yaml file in working directory, we will try to call git
	// command to full fill git info entry.
	//
	// For example, if we run main.go files in IDE or directory with command line without rk build command.
	// join the path with current working directory if user provided path is relative path
	gitConfigFile := "git.yaml"
	if !path.IsAbs(gitConfigFile) {
		wd, _ := os.Getwd()

		gitConfigFile = path.Join(wd, gitConfigFile)
	}

	if rkcommon.FileExists(gitConfigFile) {
		rkcommon.UnmarshalBootConfig("git.yaml", config)
	} else {
		config.Package, _ = rkcommon.GetPackageNameFromGitLocal()
		config.Url, _ = rkcommon.GetRemoteUrlFromGitLocal()
		config.Branch, _ = rkcommon.GetBranchFromGitLocal()
		config.Tag, _ = rkcommon.GetCurrentTagFromGitLocal()
		v, _ := rkcommon.GetLatestCommitFromGitLocal()
		config.Commit = *v
	}

	// 2: Init rk entry from config
	entry := RegisterGitInfoEntry(
		WithPackageGitInfo(config.Package),
		WithUrlGitInfo(config.Url),
		WithBranchGitInfo(config.Branch),
		WithTagGitInfo(config.Tag),
		WithCommitIdGitInfo(config.Commit.ID),
		WithCommitIdAbbrGitInfo(config.Commit.Abbr),
		WithCommitDateInfo(config.Commit.Date),
		WithCommitSubGitInfo(config.Commit.Sub),
		WithCommitterNameGitInfo(config.Commit.Committer.Name),
		WithCommitterEmailGitInfo(config.Commit.Committer.Email))

	res[GitInfoEntryName] = entry

	return res
}

// Register Entry with options.
// This function is used while creating entry from code instead of config file.
// We will override RKEntry fields if value is nil or empty if necessary.
//
// Generally, we recommend call rkctx.GlobalAppCtx.AddEntry() inside this function,
// however, we recommend to register RKEntry, ZapLoggerEntry, EventLoggerEntry with
// function of rkctx.RegisterBasicEntriesWithConfig which will register these entries to
// global context automatically.
func RegisterGitInfoEntry(opts ...GitInfoEntryOption) *GitInfoEntry {
	entry := &GitInfoEntry{
		EntryName:        GitInfoEntryName,
		EntryType:        GitInfoEntryType,
		EntryDescription: GitInfoEntryDescription,
	}

	for i := range opts {
		opts[i](entry)
	}

	GlobalAppCtx.SetGitInfoEntry(entry)

	return entry
}

// No op.
func (entry *GitInfoEntry) Bootstrap(context.Context) {
	// No op
}

// No op.
func (entry *GitInfoEntry) Interrupt(context.Context) {
	// No op
}

// Return name of entry.
func (entry *GitInfoEntry) GetName() string {
	return entry.EntryName
}

// Return type of entry.
func (entry *GitInfoEntry) GetType() string {
	return entry.EntryType
}

// Return description of entry.
func (entry *GitInfoEntry) GetDescription() string {
	return entry.EntryDescription
}

// Return string of entry.
func (entry *GitInfoEntry) String() string {
	if bytes, err := json.Marshal(entry); err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

// Construct version
func (entry *GitInfoEntry) ConstructAppVersion() string {
	version := ""

	if len(entry.Tag) > 0 {
		version = entry.Tag
	} else if len(entry.Branch) > 0 {
		if len(entry.CommitIdAbbr) > 0 {
			version = strings.Join([]string{entry.Branch, entry.CommitIdAbbr}, "-")
		} else {
			version = entry.Branch
		}
	}

	return version
}
