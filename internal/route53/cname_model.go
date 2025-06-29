package route53

import (
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type cnameModel struct {
	loc  *errorsink.Location
	name string

	pointsTo     any
	updateZoneId string
}
