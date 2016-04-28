package version

import "fmt"

const (
	BaseDefault   = "0.0.0"
	CommitDefault = "0000000000000000000000000000000000000000"
)

var (
	BaseInitial   string
	CommitInitial string
)

func Base() string {
	if BaseInitial != "" {
		return BaseInitial
	}
	return BaseDefault
}

func Commit() string {
	if CommitInitial != "" {
		return CommitInitial
	}
	return CommitDefault

}

func ShortCommit() string {
	commit := Commit()
	if len(commit) > 8 {
		return commit[:8]
	}
	return commit
}

func Short() string {
	return fmt.Sprintf("%s+%s", Base(), ShortCommit())
}

func Long() string {
	return fmt.Sprintf("%s+%s", Base(), Commit())
}
