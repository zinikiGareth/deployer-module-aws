package cfront

type websiteCoins struct {
	cachePolicy         *CachePolicyCreator
	originAccessControl *OACCreator
	rhp                 *RHPCreator
	cb                  *CacheBehaviorCreator
	distribution        *distributionCreator
}
