package main

import (
	"ziniki.org/deployer/deployer/pkg/deployer"
	"ziniki.org/deployer/modules/aws/pkg/awsmod"
)

func ProvideTestRunner(runner deployer.TestRunner) error {
	return awsmod.ProvideTestRunner(runner)
}

func RegisterWithDeployer(deployer deployer.Deployer) error {
	return awsmod.RegisterWithDeployer(deployer)
}
