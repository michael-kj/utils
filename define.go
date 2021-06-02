//go:generate  enumer -type=Env  -json -text -sql

package utils

import (
	"errors"
	"strings"
)

type Env int

const (
	Dev Env = iota + 1
	Online
	Qa
	Pl
	Local
	Unknow
)

var WrongEnvError = errors.New("wrong env value")

func EqualEnv(env interface{}, define Env) (bool, error) {
	var e Env
	switch env.(type) {
	case string:
		env, err := EnvString(strings.Title(strings.ToLower(env.(string))))
		if err != nil {
			return false, err
		}
		e = env
	case Env:
		e = env.(Env)

	}

	return e == define, nil
}
