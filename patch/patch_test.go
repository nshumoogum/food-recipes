package patch_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/nshumoogum/food-recipes/patch"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	opAdd     = "add"
	opCopy    = "copy"
	opMove    = "move"
	opRemove  = "remove"
	opReplace = "replace"
	opTest    = "test"

	opBad = "bad"

	missingPath           = "Key: 'Patch.Path' Error:Field validation for 'Path' failed on the 'required' tag"
	missingSupportedOp    = "Key: 'Patch.Op' Error:Field validation for 'Op' failed on the 'supportedops' tag"
	notPatchOp            = "Key: 'Patch.Op' Error:Field validation for 'Op' failed on the 'oneof' tag"
	missingValue          = "Key: 'Patch.Op' Error:Field validation for 'Op' failed on the 'requirevalueifopis' tag"
	missingFrom           = "Key: 'Patch.Op' Error:Field validation for 'Op' failed on the 'requirefromifopis' tag"
	sameFromAndPathValues = "Key: 'Patch.From' Error:Field validation for 'From' failed on the 'nefield' tag"

	allSupportedOps = patch.Ops{patch.OpAdd, patch.OpCopy, patch.OpMove, patch.OpRemove, patch.OpReplace, patch.OpTest}
)

func TestValidate(t *testing.T) {
	tables := []struct {
		givenTitle          string
		thenTitle           string
		patch               *patch.Patch
		supportedOps        *patch.Ops
		expectedErrorString *string
	}{
		{
			"Given a valid patch request with add operation", "Then validation is successful",
			getPatch(opAdd, "", false), &allSupportedOps, nil,
		},
		{
			"Given a valid patch request with copy operation", "Then validation is successful",
			getPatch(opCopy, "", false), &allSupportedOps, nil,
		},
		{
			"Given a valid patch request with move operation", "Then validation is successful",
			getPatch(opMove, "", false), &allSupportedOps, nil,
		},
		{
			"Given a valid patch request with remove operation", "Then validation is successful",
			getPatch(opRemove, "", false), nil, nil,
		},
		{
			"Given a valid patch request with replace operation", "Then validation is successful",
			getPatch(opReplace, "", false), nil, nil,
		},
		{
			"Given a valid patch request with test operation", "Then validation is unsuccessful",
			getPatch(opTest, "", false), nil, nil,
		},
		{
			"Given a valid patch request with replace operation", "Then validation is successful",
			getPatch(opReplace, "", false), nil, nil,
		},
		{
			"Given an invalid patch due to unsupported 'add' operation", "Then validation is unsuccessful",
			getPatch(opAdd, "", false), &patch.Ops{patch.OpRemove}, &missingSupportedOp,
		},
		{
			"Given an invalid patch due to incorrect operation", "Then validation is unsuccessful",
			getPatch(opBad, "", false), nil, &notPatchOp,
		},
		{
			"Given an invalid patch due to missing field, 'path'", "Then validation is unsuccessful",
			getPatch(opCopy, "path", false), nil, &missingPath,
		},
		{
			"Given an invalid patch due to missing field, 'value' for add operation", "Then validation is unsuccessful",
			getPatch(opAdd, "value", false), nil, &missingValue,
		},
		{
			"Given an invalid patch due to missing field, 'value' for replace operation", "Then validation is unsuccessful",
			getPatch(opReplace, "value", false), nil, &missingValue,
		},
		{
			"Given an invalid patch due to missing field, 'value' for test operation", "Then validation is unsuccessful",
			getPatch(opTest, "value", false), nil, &missingValue,
		},
		{
			"Given an invalid patch due to missing field, 'from' for copy operation", "Then validation is unsuccessful",
			getPatch(opCopy, "from", false), nil, &missingFrom,
		},
		{
			"Given an invalid patch due to missing field, 'from' for move operation", "Then validation is unsuccessful",
			getPatch(opMove, "from", false), nil, &missingFrom,
		},
		{
			"Given an invalid patch due to fields 'path' and 'from' are the same for move operation", "Then validation is unsuccessful",
			getPatch(opMove, "", true), nil, &sameFromAndPathValues,
		},
		{
			"Given an invalid patch due to fields 'path' and 'from' are the same for move operation", "Then validation is unsuccessful",
			getPatch(opMove, "", true), nil, &sameFromAndPathValues,
		},
	}

	for _, table := range tables {
		Convey(table.givenTitle, t, func() {
			Convey(table.thenTitle, func() {
				observedError := table.patch.Validate(table.supportedOps)
				if _, ok := observedError.(*validator.InvalidValidationError); ok {
					t.Fatalf("unable to register validation: %v", observedError)
				}

				if table.expectedErrorString != nil {
					So(observedError, ShouldNotBeNil)
					So(observedError.Error(), ShouldEqual, *table.expectedErrorString)
				} else {
					So(observedError, ShouldBeNil)
				}
			})
		})
	}
}

func getPatch(patchType, missingField string, fromIsSameAsPath bool) *patch.Patch {
	p := &patch.Patch{
		Op: patchType,
	}

	p.Path = "/field"

	switch patchType {
	case opAdd, opReplace, opTest:
		p.Value = "test-value"
	case opRemove:
	case opCopy, opMove:
		p.From = "/field-2"
	case opBad:
	default:
		return nil
	}

	switch missingField {
	case "path":
		p.Path = ""
	case "value":
		p.Value = nil
	case "from":
		p.From = ""
	default:
		// continue
	}

	if fromIsSameAsPath {
		p.From = p.Path
	}

	return p
}
