package appverx

import (
	_ "embed"
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
)

type Core struct {
	Major uint
	Minor uint
	Patch uint
}
type PreRelease string
type Build struct {
	Commit string
	Dirty  bool
}

type SemVer struct {
	Core       Core
	PreRelease PreRelease
	Build      Build
}

type Info struct {
	Version SemVer
	Time    string
}

var semverCoreRe = regexp.MustCompile(
	`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`,
)

func NewInfo(c string) (Info, error) {
	v, err := ParseSemVerCore(c)
	if err != nil {
		return Info{}, err
	}

	i := Info{
		Version: v,
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return i, nil
	}

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			i.Version.Build.Commit = s.Value
		case "vcs.modified":
			i.Version.Build.Dirty = s.Value == "true"
		case "vcs.time":
			i.Time = s.Value
		}
	}

	return i, nil
}

func (c Core) String() string {
	return fmt.Sprintf("%d.%d.%d", c.Major, c.Minor, c.Patch)
}

func (p PreRelease) HasValue() bool {
	return len(p) > 0
}

func (p PreRelease) String() string {
	return string(p)
}

func (b Build) ShortCommit() Build {
	c := b.Commit
	if len(c) > 7 {
		c = c[:7]
	}

	return Build{
		Commit: c,
		Dirty:  b.Dirty,
	}
}

func (b Build) HasCommit() bool {
	return b.Commit != ""
}

func (b Build) String() string {
	if b.Commit == "" {
		return ""
	}

	if b.Dirty {
		return b.Commit + ".dirty"
	}
	return b.Commit
}

func (v SemVer) String() string {
	s := fmt.Sprint(v.Core)
	if v.PreRelease.HasValue() {
		s += fmt.Sprint("-", v.PreRelease)
	}
	if v.Build.HasCommit() {
		s += fmt.Sprint("+git.", v.Build)
	}

	return s
}

func (i Info) String() string {
	v := i.Version
	v.Build = Build{}

	s := fmt.Sprint(v, " build: ")

	c := "unknown"
	if i.Version.Build.HasCommit() {
		c = fmt.Sprint(i.Version.Build)
	}
	s += c

	if i.Time != "" {
		s += fmt.Sprint(" (", i.Time, ")")
	}

	return s
}

func mustUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

func ParseSemVerCore(c string) (SemVer, error) {
	m := semverCoreRe.FindStringSubmatch(c)
	if m == nil {
		return SemVer{}, fmt.Errorf("invalid SemVer: %s", c)
	}

	return SemVer{
		Core: Core{
			Major: mustUint(m[1]),
			Minor: mustUint(m[2]),
			Patch: mustUint(m[3]),
		},
		PreRelease: PreRelease(m[4]),
	}, nil
}
