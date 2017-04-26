package environment

func NewDefaultReporter(prefix string) (Reporter, error) {
	return NewReporter(GetValue("ENV", prefix), prefix)
}
