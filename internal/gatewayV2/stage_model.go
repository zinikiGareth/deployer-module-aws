package gatewayV2

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type StageModel struct {
	name string
	loc  *errorsink.Location
	coin corebottom.CoinId

	api fmt.Stringer
}

type StageAWSModel struct {
	name string
	// coin corebottom.CoinId
}
