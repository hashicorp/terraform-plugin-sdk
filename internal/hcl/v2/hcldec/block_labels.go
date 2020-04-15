package hcldec

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/hcl/v2"
)

type blockLabel struct {
	Value string
	Range hcl.Range
}

func labelsForBlock(block *hcl.Block) []blockLabel {
	ret := make([]blockLabel, len(block.Labels))
	for i := range block.Labels {
		ret[i] = blockLabel{
			Value: block.Labels[i],
			Range: block.LabelRanges[i],
		}
	}
	return ret
}
