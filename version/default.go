package version

// WARNING: Concurrent modification of these global variables is unsupported

var (
	Base        string
	ShortCommit string
	FullCommit  string
)

func NewDefaultReporter() (Reporter, error) {
	return NewReporter(Base, ShortCommit, FullCommit)
}
