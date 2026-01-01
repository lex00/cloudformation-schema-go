package spec

import (
	"strings"
)

// GetResourceType returns the resource type definition for the given type name.
// Returns nil if the resource type is not found.
func (s *Spec) GetResourceType(typeName string) *ResourceType {
	if rt, ok := s.ResourceTypes[typeName]; ok {
		return &rt
	}
	return nil
}

// GetPropertyType returns the property type definition for the given type name.
// Property type names are typically "AWS::Service::Resource.PropertyName".
// Returns nil if the property type is not found.
func (s *Spec) GetPropertyType(typeName string) *PropertyType {
	if pt, ok := s.PropertyTypes[typeName]; ok {
		return &pt
	}
	return nil
}

// HasResourceType returns true if the spec contains the given resource type.
func (s *Spec) HasResourceType(typeName string) bool {
	_, ok := s.ResourceTypes[typeName]
	return ok
}

// HasPropertyType returns true if the spec contains the given property type.
func (s *Spec) HasPropertyType(typeName string) bool {
	_, ok := s.PropertyTypes[typeName]
	return ok
}

// ResourceTypeNames returns all resource type names in the spec.
func (s *Spec) ResourceTypeNames() []string {
	names := make([]string, 0, len(s.ResourceTypes))
	for name := range s.ResourceTypes {
		names = append(names, name)
	}
	return names
}

// PropertyTypeNames returns all property type names in the spec.
func (s *Spec) PropertyTypeNames() []string {
	names := make([]string, 0, len(s.PropertyTypes))
	for name := range s.PropertyTypes {
		names = append(names, name)
	}
	return names
}

// GetRequiredProperties returns a list of required property names for the resource type.
func (rt *ResourceType) GetRequiredProperties() []string {
	var required []string
	for name, prop := range rt.Properties {
		if prop.Required {
			required = append(required, name)
		}
	}
	return required
}

// GetProperty returns the property definition for the given property name.
// Returns nil if the property is not found.
func (rt *ResourceType) GetProperty(name string) *Property {
	if prop, ok := rt.Properties[name]; ok {
		return &prop
	}
	return nil
}

// HasProperty returns true if the resource type has the given property.
func (rt *ResourceType) HasProperty(name string) bool {
	_, ok := rt.Properties[name]
	return ok
}

// GetAttribute returns the attribute definition for the given attribute name.
// Returns nil if the attribute is not found.
func (rt *ResourceType) GetAttribute(name string) *Attribute {
	if attr, ok := rt.Attributes[name]; ok {
		return &attr
	}
	return nil
}

// HasAttribute returns true if the resource type has the given attribute.
func (rt *ResourceType) HasAttribute(name string) bool {
	_, ok := rt.Attributes[name]
	return ok
}

// AttributeNames returns all attribute names for the resource type.
func (rt *ResourceType) AttributeNames() []string {
	names := make([]string, 0, len(rt.Attributes))
	for name := range rt.Attributes {
		names = append(names, name)
	}
	return names
}

// GetProperty returns the property definition for the given property name.
// Returns nil if the property is not found.
func (pt *PropertyType) GetProperty(name string) *Property {
	if prop, ok := pt.Properties[name]; ok {
		return &prop
	}
	return nil
}

// HasProperty returns true if the property type has the given property.
func (pt *PropertyType) HasProperty(name string) bool {
	_, ok := pt.Properties[name]
	return ok
}

// GetRequiredProperties returns a list of required property names for the property type.
func (pt *PropertyType) GetRequiredProperties() []string {
	var required []string
	for name, prop := range pt.Properties {
		if prop.Required {
			required = append(required, name)
		}
	}
	return required
}

// IsPrimitive returns true if the property is a primitive type.
func (p *Property) IsPrimitive() bool {
	return p.PrimitiveType != ""
}

// IsList returns true if the property is a list type.
func (p *Property) IsList() bool {
	return p.Type == "List"
}

// IsMap returns true if the property is a map type.
func (p *Property) IsMap() bool {
	return p.Type == "Map"
}

// IsComplex returns true if the property is a complex (nested) type.
func (p *Property) IsComplex() bool {
	return p.Type != "" && p.Type != "List" && p.Type != "Map"
}

// GetPropertyTypeForResource returns the full property type name for a nested property.
// For example, given resourceType "AWS::S3::Bucket" and property type "CorsConfiguration",
// it returns "AWS::S3::Bucket.CorsConfiguration".
func GetPropertyTypeForResource(resourceType, propertyType string) string {
	return resourceType + "." + propertyType
}

// ParsePropertyTypeName splits a property type name into resource type and property name.
// For example, "AWS::S3::Bucket.CorsConfiguration" returns ("AWS::S3::Bucket", "CorsConfiguration").
func ParsePropertyTypeName(fullName string) (resourceType, propertyName string) {
	parts := strings.SplitN(fullName, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return fullName, ""
}
