package route53

import (
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type aliasModel struct {
	loc  *errorsink.Location
	name string

	otherDomain  any
	updateZoneId string
	aliasZoneId  string
}
