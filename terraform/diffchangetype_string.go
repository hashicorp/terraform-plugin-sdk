// Code generated by "stringer -type DiffChangeType"; DO NOT EDIT.

package terraform

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[DiffInvalid-0]
	_ = x[DiffNone-1]
	_ = x[DiffCreate-2]
	_ = x[DiffUpdate-3]
	_ = x[DiffDestroy-4]
	_ = x[DiffDestroyCreate-5]
	_ = x[DiffRefresh-6]
}

const _DiffChangeType_name = "DiffInvalidDiffNoneDiffCreateDiffUpdateDiffDestroyDiffDestroyCreateDiffRefresh"

var _DiffChangeType_index = [...]uint8{0, 11, 19, 29, 39, 50, 67, 78}

func (i DiffChangeType) String() string {
	if i >= DiffChangeType(len(_DiffChangeType_index)-1) {
		return "DiffChangeType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _DiffChangeType_name[_DiffChangeType_index[i]:_DiffChangeType_index[i+1]]
}
