package cfront

import (
	"fmt"

	"ziniki.org/deployer/driver/pkg/errorsink"
)

type invalidateModel struct {
	loc      *errorsink.Location
	distroId string
	paths    []fmt.Stringer
}
