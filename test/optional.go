package test

import "github.com/tidepool-org/platform/pointer"

type Option struct {
	allowOptional *bool
}

func (o Option) AllowOptional() bool {
	return o.allowOptional != nil && *o.allowOptional
}

func AllowOptional() Option {
	return Option{
		allowOptional: pointer.From(true),
	}
}

func Options(options []Option) Option {
	option := Option{}
	for _, o := range options {
		if o.allowOptional != nil {
			option.allowOptional = o.allowOptional
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

func IsConditionallyTrue(options ...Option) bool {
	return !Options(options).AllowOptional() || RandomBool()
}

func RandomOptional[T any](generator func() T, options ...Option) *T {
	return Conditional(generator, IsConditionallyTrue(options...))
}

func RandomOptionalWithOptions[T any](generator func(options ...Option) T, options ...Option) *T {
	return Conditional(optionsAdapter(generator, options...), IsConditionallyTrue(options...))
}

func RandomOptionalPointer[T any](generator func() *T, options ...Option) *T {
	return ConditionalPointer(generator, IsConditionallyTrue(options...))
}

func RandomOptionalPointerWithOptions[T any](generator func(options ...Option) *T, options ...Option) *T {
	return ConditionalPointer(optionsAdapter(generator, options...), IsConditionallyTrue(options...))
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
