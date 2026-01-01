package spec

// Spec represents the CloudFormation Resource Specification.
type Spec struct {
	ResourceSpecificationVersion string                  `json:"ResourceSpecificationVersion"`
	ResourceTypes                map[string]ResourceType `json:"ResourceTypes"`
	PropertyTypes                map[string]PropertyType `json:"PropertyTypes"`
}

// ResourceType is a CloudFormation resource type definition.
type ResourceType struct {
	Documentation        string               `json:"Documentation"`
	Attributes           map[string]Attribute `json:"Attributes"`
	Properties           map[string]Property  `json:"Properties"`
	AdditionalProperties bool                 `json:"AdditionalProperties"`
}

// PropertyType is a property type definition (nested structures).
type PropertyType struct {
	Documentation string              `json:"Documentation"`
	Properties    map[string]Property `json:"Properties"`
}

// Property is a property definition.
type Property struct {
	Documentation     string `json:"Documentation"`
	Required          bool   `json:"Required"`
	PrimitiveType     string `json:"PrimitiveType"`     // String, Integer, Boolean, etc.
	Type              string `json:"Type"`              // List, Map, or property type name
	ItemType          string `json:"ItemType"`          // For List/Map
	PrimitiveItemType string `json:"PrimitiveItemType"` // For List/Map of primitives
	UpdateType        string `json:"UpdateType"`        // Mutable, Immutable, Conditional
	DuplicatesAllowed bool   `json:"DuplicatesAllowed"`
}

// Attribute is a resource attribute (for GetAtt).
type Attribute struct {
	PrimitiveType     string `json:"PrimitiveType"`
	Type              string `json:"Type"`
	PrimitiveItemType string `json:"PrimitiveItemType"`
	ItemType          string `json:"ItemType"`
}
