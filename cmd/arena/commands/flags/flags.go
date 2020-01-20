package flags

import (
	flag "github.com/spf13/pflag"
	"strconv"
	"time"
)

type intNullableArgument struct {
	variableToSet **int
}

func newIntNullableArgument(variable **int) *intNullableArgument {
	return &intNullableArgument{
		variableToSet: variable,
	}
}

func (argument *intNullableArgument) Set(s string) error {
	v, err := strconv.Atoi(s)
	*argument.variableToSet = &v
	return err
}

func (argument *intNullableArgument) Type() string {
	return "int"
}

func (argument *intNullableArgument) String() string {
	if *argument.variableToSet == nil {
		return ""
	}

	return strconv.Itoa(**argument.variableToSet)
}

func AddIntNullableFlagP(f *flag.FlagSet, variableToSet **int, name string, shorthand string, usage string) {
	f.VarPF(newIntNullableArgument(variableToSet), name, shorthand, usage)
}

func AddIntNullableFlag(f *flag.FlagSet, variableToSet **int, name string, usage string) {
	AddIntNullableFlagP(f, variableToSet, name, "", usage)
}

type boolNullableArgument struct {
	variableToSet **bool
}

func newBoolNullableArgument(variable **bool) *boolNullableArgument {
	return &boolNullableArgument{
		variableToSet: variable,
	}
}

func (argument *boolNullableArgument) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*argument.variableToSet = &v
	return err
}

func (argument *boolNullableArgument) Type() string {
	return "bool"
}

func (argument *boolNullableArgument) String() string {
	if *argument.variableToSet == nil {
		return ""
	}

	return strconv.FormatBool(**argument.variableToSet)
}

func AddBoolNullableFlagP(f *flag.FlagSet, variableToSet **bool, name string, shorthand string, usage string) {
	flag := f.VarPF(newBoolNullableArgument(variableToSet), name, shorthand, usage)
	flag.NoOptDefVal = "true"
}

func AddBoolNullableFlag(f *flag.FlagSet, variableToSet **bool, name string, usage string) {
	AddBoolNullableFlagP(f, variableToSet, name, "", usage)
}

type durationNullableArgument struct {
	variableToSet **time.Duration
}

func newDurationNullableArgument(variable **time.Duration) *durationNullableArgument {
	return &durationNullableArgument{
		variableToSet: variable,
	}
}

func (argument *durationNullableArgument) Set(s string) error {
	v, err := time.ParseDuration(s)
	*argument.variableToSet = &v
	return err
}

func (argument *durationNullableArgument) Type() string {
	return "duration"
}

func (argument *durationNullableArgument) String() string {
	if *argument.variableToSet == nil {
		return ""
	}

	return (**argument.variableToSet).String()
}

func AddDurationNullableFlagP(f *flag.FlagSet, variableToSet **time.Duration, name string, shorthand string, usage string) {
	f.VarP(newDurationNullableArgument(variableToSet), name, shorthand, usage)
}
