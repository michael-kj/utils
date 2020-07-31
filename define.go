//go:generate  enumer -type=Env  -json -text -sql

package utils

import "errors"

type Env int

const (
	Dev Env = iota + 1
	Online
	Qa
	Pl
	Local
)

var WrongEnvError = errors.New("wrong env value")

func CheckEnv(env Env) error {
	if !env.IsAEnv() {
		return WrongEnvError
	}
	return nil
}
