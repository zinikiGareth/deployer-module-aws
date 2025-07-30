package lambda

import (
	"fmt"

	"ziniki.org/deployer/driver/pkg/driverbottom"
)

type publishVersionAWS struct {
	functionName string
	aliasName    string

	publishedVersion string
	aliasVersion     string
	aliasRevId       string
}

type publishVersionModel struct {
	publish driverbottom.AsNumber
	asAlias fmt.Stringer

	name fmt.Stringer
}
