package main

import (
	"os"

	"ziniki.org/deployer/coremod/pkg/coretop"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/modules/aws/pkg/awsmod"
)

func main() {
	if len(os.Args) == 1 {
		drivertop.Usage()
		os.Exit(1)
	}

	stat := drivertop.RunDeployerWithConfig(loadMods, os.Args[1:])
	os.Exit(stat)
}

func loadMods(driver driverbottom.Driver) error {
	if err := coretop.RegisterWithDriver(driver); err != nil {
		return err
	}
	return awsmod.RegisterWithDriver(driver)
}
