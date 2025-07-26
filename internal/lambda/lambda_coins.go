package lambda

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/modules/aws/internal/iam"
)

type lambdaCoins struct {
	withRole    *iam.WithRole
	roleCreator corebottom.Ensurable
	lambda      *lambdaCreator
}
