package target

import "infralog/tfstate"

type Target interface {
	Write(*tfstate.StateDiff) error
}
