package patch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Possible patch operations
const (
	OpAdd Operation = iota
	OpCopy
	OpMove
	OpRemove
	OpReplace
	OpTest

	value = "value"
	from  = "from"
)

var validOps = []string{"add", "copy", "move", "remove", "replace", "test"}

// Ops - a list of patch operations
type Ops []Operation

// Operation - iota enum of possible patch operations
type Operation int

// Patch models an HTTP patch operation request, according to RFC 6902
type Patch struct {
	Op    string      `json:"op"                validate:"required,supportedops,oneof=add copy move remove replace test,requirevalueifopis=add replace test,requirefromifopis=copy move"`
	Path  string      `json:"path"              validate:"required"`
	From  string      `json:"from,omitempty"    validate:"nefield=Path"`
	Value interface{} `json:"value,omitempty"`
}

func (op Operation) String() string {
	return validOps[op]
}

func (ops Ops) StringSlice() []string {
	opsStringSlice := []string{}
	for _, op := range ops {
		opsStringSlice = append(opsStringSlice, op.String())
	}
	return opsStringSlice
}

// Get gets a list of patches from the request body and returns it in the form of []Patch
func Get(ctx context.Context, requestBody io.ReadCloser) ([]byte, *[]Patch, error) {
	b, err := io.ReadAll(requestBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read and get patch request body")
	}

	if len(b) == 0 {
		return nil, nil, fmt.Errorf("empty request body given")
	}

	patches := []Patch{}
	err = json.Unmarshal(b, &patches)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal patch request body")
	}

	if len(patches) < 1 {
		return nil, nil, fmt.Errorf("no patches given in request body")
	}

	return b, &patches, nil
}

// Validate against patch object to check fundamental data of the patch object
func (p *Patch) Validate(supportedOps *Ops) error {
	validate := validator.New()

	fmt.Printf("Patch is: %v", p)

	// Validate possible operations
	if err := validate.RegisterValidation("supportedops", getOpsValidator(supportedOps, p)); err != nil {
		return fmt.Errorf("failed to register ops validator: %w", err)
	}

	if err := validate.RegisterValidation("requirevalueifopis", getRequiredFieldIfOperationIsValidator(p, value)); err != nil {
		return fmt.Errorf("failed to register requirevalueifopis validator: %w", err)
	}

	if err := validate.RegisterValidation("requirefromifopis", getRequiredFieldIfOperationIsValidator(p, from)); err != nil {
		return fmt.Errorf("failed to register requirevalueifopis validator: %w", err)
	}

	return validate.Struct(p)
}

func getRequiredFieldIfOperationIsValidator(p *Patch, field string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		// retrieve operation values that would require this field
		vals := parseOneOfParam2(fl.Param())

		for i := range vals {
			if vals[i] == p.Op {
				switch field {
				case value:
					if p.Value == nil {
						return false
					}
				case from:
					if p.From == "" {
						return false
					}
				}
				break
			}
		}

		return true
	}
}

var splitParamsRegex = regexp.MustCompile(`'[^']*'|\S+`)

func parseOneOfParam2(s string) []string {
	oneofValsCache := map[string][]string{}

	vals, ok := oneofValsCache[s]
	if !ok {
		vals = splitParamsRegex.FindAllString(s, -1)
		for i := 0; i < len(vals); i++ {
			vals[i] = strings.Replace(vals[i], "'", "", -1)
		}
		oneofValsCache[s] = vals
	}

	return vals
}

func getOpsValidator(supportedOps *Ops, p *Patch) validator.Func {
	return func(fl validator.FieldLevel) bool {
		return p.isOpSupported(supportedOps)
	}
}

// isOpSupported checks that the patch op is in the provided list of supported Ops
func (p *Patch) isOpSupported(supportedOps *Ops) bool {
	if supportedOps == nil {
		return true
	}

	for _, op := range *supportedOps {
		if p.Op == op.String() {
			return true
		}
	}
	return false
}
