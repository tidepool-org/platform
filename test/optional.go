package test

import "github.com/tidepool-org/platform/pointer"

const (
	OptionalsRequire = "require"
	OptionalsAllow   = "allow"
	OptionalsForbid  = "forbid"
)

func Optionals() []string {
	return []string{
		OptionalsRequire,
		OptionalsAllow,
		OptionalsForbid,
	}
}

type Option struct {
	optionals *string
}

func (o Option) AreOptionalsRequired() bool {
	return o.optionals == nil || *o.optionals == OptionalsRequire
}

func (o Option) AreOptionalsAllowed() bool {
	return o.optionals != nil && *o.optionals == OptionalsAllow
}

func (o Option) AreOptionalsForbidden() bool {
	return o.optionals != nil && *o.optionals == OptionalsForbid
}

func RandomOptionals() Option {
	return Option{
		optionals: pointer.From(RandomStringFromArray(Optionals())),
	}
}

func RequireOptionals() Option {
	return Option{
		optionals: pointer.From(OptionalsRequire),
	}
}

func AllowOptionals() Option {
	return Option{
		optionals: pointer.From(OptionalsAllow),
	}
}

func ForbidOptionals() Option {
	return Option{
		optionals: pointer.From(OptionalsForbid),
	}
}

func Options(options []Option) Option {
	option := Option{}
	for _, o := range options {
		if o.optionals != nil {
			option.optionals = o.optionals
		}
	}
	return option
}

func Conditional[T any](generator func() T, condition bool) *T {
	if condition {
		return pointer.From(generator())
	} else {
		return nil
	}
}

func ConditionalPointer[T any](generator func() *T, condition bool) *T {
	if condition {
		return generator()
	} else {
		return nil
	}
}

func IsOptionalPresent(options ...Option) bool {
	resolvedOptions := Options(options)
	return resolvedOptions.AreOptionalsRequired() || (resolvedOptions.AreOptionalsAllowed() && RandomBool())
}

func RandomOptional[T any](generator func() T, options ...Option) *T {
	return Conditional(generator, IsOptionalPresent(options...))
}

func RandomOptionalWithOptions[T any](generator func(options ...Option) T, options ...Option) *T {
	return Conditional(optionsAdapter(generator, options...), IsOptionalPresent(options...))
}

func RandomOptionalPointer[T any](generator func() *T, options ...Option) *T {
	return ConditionalPointer(generator, IsOptionalPresent(options...))
}

func RandomOptionalPointerWithOptions[T any](generator func(options ...Option) *T, options ...Option) *T {
	return ConditionalPointer(optionsAdapter(generator, options...), IsOptionalPresent(options...))
}

func Constant[T any](value T) func() T {
	return func() T { return value }
}

func ConstantPointer[T any](value T) func() *T {
	return func() *T { return pointer.From(value) }
}

func optionsAdapter[T any](generator func(options ...Option) T, options ...Option) func() T {
	return func() T { return generator(options...) }
}
