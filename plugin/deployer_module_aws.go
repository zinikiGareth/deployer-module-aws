package main

import (
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/modules/aws/pkg/awsmod"
)

func ProvideTestRunner(runner driverbottom.TestRunner) error {
	return awsmod.ProvideTestRunner(runner)
}

func RegisterWithDriver(deployer driverbottom.Driver) error {
	return awsmod.RegisterWithDriver(deployer)
}
