package lambda

import (
	"fmt"

	"ziniki.org/deployer/driver/pkg/driverbottom"
)

type publishVersionAWS struct {
	publishedVersion string
	aliasVersion     string
	aliasRevId       string
}

type publishVersionModel struct {
	publish driverbottom.AsNumber
	asAlias fmt.Stringer

	name fmt.Stringer
}
