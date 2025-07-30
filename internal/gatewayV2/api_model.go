package gatewayV2

import (
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type ApiAWSModel struct {
	api *types.Api
}

type ApiModel struct {
	name     string
	loc      *errorsink.Location
	coin     corebottom.CoinId
	protocol types.ProtocolType
}
