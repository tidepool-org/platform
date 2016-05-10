package version

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

type Version interface {
	Base() string
	ShortCommit() string
	FullCommit() string
	Short() string
	Long() string
}

func Current() Version {
	return _current
}
