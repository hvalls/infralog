package target

import (
	"infralog/config"
	"infralog/tfstate"
)

type Target interface {
	Write(*tfstate.StateDiff, config.TFState) error
}
