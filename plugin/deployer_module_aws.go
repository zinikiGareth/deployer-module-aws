package main

import (
	"ziniki.org/deployer/deployer/pkg/deployer"
	"ziniki.org/deployer/modules/aws/pkg/awsmod"
)

func ProvideTestRunner(runner deployer.TestRunner) error {
	return awsmod.ProvideTestRunner(runner)
}

func RegisterWithDriver(deployer deployer.Driver) error {
	return awsmod.RegisterWithDriver(deployer)
}
