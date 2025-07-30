package lambda

import (
	"fmt"

	"ziniki.org/deployer/driver/pkg/driverbottom"
)

type publishVersionModel struct {
	publish driverbottom.AsNumber
	asAlias fmt.Stringer

	name fmt.Stringer
}
