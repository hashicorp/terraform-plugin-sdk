// Copyright IBM Corp. 2019, 2026
// SPDX-License-Identifier: MPL-2.0

package structure

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SuppressJsonDiff(k, oldValue, newValue string, d *schema.ResourceData) bool {
	var o, n interface{}
	if err := json.Unmarshal([]byte(oldValue), &o); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(newValue), &n); err != nil {
		return false
	}
	return reflect.DeepEqual(o, n)
}
