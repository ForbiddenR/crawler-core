package utils

import "github.com/crawlab-team/crawlab-core/interfaces"

func GetUserFromArgs(args ...any) interfaces.User {
	for _, arg := range args {
		switch t := arg.(type) {
		case interfaces.User:
			return t
		}
	}
	return nil
}
