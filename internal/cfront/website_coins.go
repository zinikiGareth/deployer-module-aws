package cfront

type websiteCoins struct {
	cachePolicy         *CachePolicyCreator
	originAccessControl *OACCreator
	cbs                 []*CacheBehaviorCreator
	distribution        *distributionCreator
}
