package gatewayV2

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type IntegrationAWSModel struct {
	integration *types.Integration
}

type IntegrationModel struct {
	name string
	loc  *errorsink.Location
	coin corebottom.CoinId

	api    fmt.Stringer
	region fmt.Stringer
	itype  fmt.Stringer
	uri    fmt.Stringer
}
