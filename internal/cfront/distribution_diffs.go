package cfront

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
)

// In attempting to use this as a springboard for the "general" case, I think I have over-engineered this.
// Not as much as I intended to, when I was going to use an "action" metaphor, which I think would be needed if
// you want to create & delete actual objects.

// But I think you can just take the []*cbModel from desired (cbs below) and just shove it in the insert clause

type distributionDiffs struct {
	models []*cbModel
}

func figureDiffs(tools *corebottom.Tools, found, desired *DistributionModel) *distributionDiffs {
	foo := desired.behaviors.Eval(tools.Storage)
	cbs := foo.([]interface{})
	doSomething := false
	diffs := &distributionDiffs{}
	for i, cbi := range cbs {
		cb := cbi.(*cbModel)
		if i < len(found.foundBehaviors) && found.foundBehaviors[i].name == cb.name {
			diffs.models = append(diffs.models, found.foundBehaviors[i])
			continue
		}
		doSomething = true
		moved := false
		for j := 0; j < len(found.foundBehaviors); j++ {
			if found.foundBehaviors[j].name == cb.name {
				diffs.models = append(diffs.models, found.foundBehaviors[j])
				moved = true
			}
		}
		if !moved {
			diffs.models = append(diffs.models, cb)
		}
	}
	if doSomething {
		return diffs
	} else {
		return nil
	}
}

func (diffs *distributionDiffs) apply(_ *corebottom.Tools, _ *cloudfront.Client, created *DistributionModel) *types.CacheBehaviors {
	created.foundBehaviors = diffs.models
	cbs := []types.CacheBehavior{}
	for _, cbc := range diffs.models {
		resolved := cbc.Complete()
		cbs = append(cbs, resolved)
	}

	cbl := int32(len(cbs))
	return &types.CacheBehaviors{Quantity: &cbl, Items: cbs}
}
