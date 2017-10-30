package version

type Reporter interface {
	Base() string
	ShortCommit() string
	FullCommit() string
	Short() string
	Long() string
}
