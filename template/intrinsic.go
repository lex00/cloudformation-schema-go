package template

// IntrinsicType represents a CloudFormation intrinsic function type.
type IntrinsicType int

const (
	IntrinsicRef IntrinsicType = iota
	IntrinsicGetAtt
	IntrinsicSub
	IntrinsicJoin
	IntrinsicSelect
	IntrinsicGetAZs
	IntrinsicIf
	IntrinsicEquals
	IntrinsicAnd
	IntrinsicOr
	IntrinsicNot
	IntrinsicCondition
	IntrinsicFindInMap
	IntrinsicBase64
	IntrinsicCidr
	IntrinsicImportValue
	IntrinsicSplit
	IntrinsicTransform
	IntrinsicValueOf
)

// String returns the CloudFormation name for this intrinsic type.
func (t IntrinsicType) String() string {
	switch t {
	case IntrinsicRef:
		return "Ref"
	case IntrinsicGetAtt:
		return "GetAtt"
	case IntrinsicSub:
		return "Sub"
	case IntrinsicJoin:
		return "Join"
	case IntrinsicSelect:
		return "Select"
	case IntrinsicGetAZs:
		return "GetAZs"
	case IntrinsicIf:
		return "If"
	case IntrinsicEquals:
		return "Equals"
	case IntrinsicAnd:
		return "And"
	case IntrinsicOr:
		return "Or"
	case IntrinsicNot:
		return "Not"
	case IntrinsicCondition:
		return "Condition"
	case IntrinsicFindInMap:
		return "FindInMap"
	case IntrinsicBase64:
		return "Base64"
	case IntrinsicCidr:
		return "Cidr"
	case IntrinsicImportValue:
		return "ImportValue"
	case IntrinsicSplit:
		return "Split"
	case IntrinsicTransform:
		return "Transform"
	case IntrinsicValueOf:
		return "ValueOf"
	default:
		return "Unknown"
	}
}

// Intrinsic represents a parsed CloudFormation intrinsic function.
// The Args structure varies by intrinsic type:
//   - Ref: string (logical_id)
//   - GetAtt: []string{logical_id, attribute}
//   - Sub: string or []any{template, variables_map}
//   - Join: []any{delimiter, values_list}
//   - Select: []any{index, list}
//   - If: []any{condition_name, true_value, false_value}
//   - Equals: []any{value1, value2}
//   - And/Or: []any (list of conditions)
//   - Not: any (single condition)
//   - FindInMap: []any{map_name, top_key, second_key}
//   - Base64: any (value to encode)
//   - Cidr: []any{ip_block, count, cidr_bits}
//   - ImportValue: any (export name)
//   - Split: []any{delimiter, source}
//   - GetAZs: string (region, empty for current)
type Intrinsic struct {
	Type IntrinsicType
	Args any
}
