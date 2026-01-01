package template

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseTemplate parses a CloudFormation template file into a Template.
// Supports both YAML and JSON formats.
func ParseTemplate(path string) (*Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	return ParseTemplateContent(content, path)
}

// ParseTemplateContent parses CloudFormation template content into a Template.
func ParseTemplateContent(content []byte, sourceName string) (*Template, error) {
	contentStr := string(content)

	// Check for unsupported custom tags
	if strings.Contains(contentStr, "!Rain::") {
		return nil, fmt.Errorf("template uses Rain-specific tags (!Rain::S3, etc.) which are not standard CloudFormation")
	}

	// Check for Kubernetes manifests
	if strings.Contains(contentStr, "apiVersion:") && strings.Contains(contentStr, "kind:") {
		return nil, fmt.Errorf("file appears to be a Kubernetes manifest, not a CloudFormation template")
	}

	// Try YAML first with custom node handling
	var rootNode yaml.Node
	err := yaml.Unmarshal(content, &rootNode)
	if err != nil {
		// Try JSON
		var data map[string]any
		err = json.Unmarshal(content, &data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template as YAML or JSON: %w", err)
		}
		return parseFromMap(data, sourceName)
	}

	// Parse from YAML node tree to handle tags
	data := parseYAMLNode(&rootNode)
	if m, ok := data.(map[string]any); ok {
		return parseFromMap(m, sourceName)
	}

	return nil, fmt.Errorf("template root must be a mapping")
}

// parseYAMLNode recursively converts a yaml.Node to Go values, handling CF intrinsic tags.
func parseYAMLNode(node *yaml.Node) any {
	return parseYAMLNodeWithVisited(node, make(map[*yaml.Node]bool))
}

// parseYAMLNodeWithVisited is the internal implementation with cycle detection.
func parseYAMLNodeWithVisited(node *yaml.Node, visited map[*yaml.Node]bool) any {
	if node == nil {
		return nil
	}

	// Cycle detection
	if visited[node] {
		return nil // Break cycle
	}
	visited[node] = true

	// Handle document node
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) > 0 {
			return parseYAMLNodeWithVisited(node.Content[0], visited)
		}
		return nil
	}

	// Check for CloudFormation intrinsic function tags (single !, not !! standard tags)
	if node.Tag != "" && strings.HasPrefix(node.Tag, "!") && !strings.HasPrefix(node.Tag, "!!") {
		return parseIntrinsicTagWithVisited(node, visited)
	}

	switch node.Kind {
	case yaml.ScalarNode:
		return parseScalar(node)

	case yaml.SequenceNode:
		result := make([]any, 0, len(node.Content))
		for _, child := range node.Content {
			result = append(result, parseYAMLNodeWithVisited(child, visited))
		}
		return result

	case yaml.MappingNode:
		result := make(map[string]any)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			key := parseScalarString(keyNode)
			result[key] = parseYAMLNodeWithVisited(valueNode, visited)
		}
		return result

	case yaml.AliasNode:
		return parseYAMLNodeWithVisited(node.Alias, visited)
	}

	return nil
}

func parseScalar(node *yaml.Node) any {
	var value any
	if err := node.Decode(&value); err != nil {
		return node.Value
	}
	return value
}

func parseScalarString(node *yaml.Node) string {
	if node.Kind == yaml.ScalarNode {
		return node.Value
	}
	return ""
}

// parseNodeContentsWithVisited parses the contents of a tagged node without re-checking the tag.
// This prevents infinite recursion when an intrinsic like !Base64 wraps another structure.
func parseNodeContentsWithVisited(node *yaml.Node, visited map[*yaml.Node]bool) any {
	switch node.Kind {
	case yaml.ScalarNode:
		return parseScalar(node)
	case yaml.SequenceNode:
		result := make([]any, 0, len(node.Content))
		for _, child := range node.Content {
			result = append(result, parseYAMLNodeWithVisited(child, visited))
		}
		return result
	case yaml.MappingNode:
		result := make(map[string]any)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			key := parseScalarString(keyNode)
			result[key] = parseYAMLNodeWithVisited(valueNode, visited)
		}
		return result
	}
	return nil
}

// parseIntrinsicTagWithVisited handles CloudFormation intrinsic function YAML tags.
func parseIntrinsicTagWithVisited(node *yaml.Node, visited map[*yaml.Node]bool) *Intrinsic {
	tag := strings.TrimPrefix(node.Tag, "!")

	switch tag {
	case "Ref":
		return &Intrinsic{Type: IntrinsicRef, Args: parseScalarString(node)}

	case "GetAtt":
		if node.Kind == yaml.ScalarNode {
			// !GetAtt Resource.Attribute format
			parts := strings.SplitN(node.Value, ".", 2)
			return &Intrinsic{Type: IntrinsicGetAtt, Args: parts}
		}
		// !GetAtt [Resource, Attribute] format
		if node.Kind == yaml.SequenceNode && len(node.Content) >= 2 {
			parts := make([]string, len(node.Content))
			for i, child := range node.Content {
				parts[i] = parseScalarString(child)
			}
			return &Intrinsic{Type: IntrinsicGetAtt, Args: parts}
		}

	case "Sub":
		if node.Kind == yaml.ScalarNode {
			return &Intrinsic{Type: IntrinsicSub, Args: node.Value}
		}
		if node.Kind == yaml.SequenceNode {
			args := make([]any, 0, len(node.Content))
			for _, child := range node.Content {
				args = append(args, parseYAMLNodeWithVisited(child, visited))
			}
			return &Intrinsic{Type: IntrinsicSub, Args: args}
		}

	case "Join":
		if node.Kind == yaml.SequenceNode && len(node.Content) >= 2 {
			delimiter := parseScalarString(node.Content[0])
			values := parseYAMLNodeWithVisited(node.Content[1], visited)
			return &Intrinsic{Type: IntrinsicJoin, Args: []any{delimiter, values}}
		}

	case "Select":
		if node.Kind == yaml.SequenceNode && len(node.Content) >= 2 {
			index := parseYAMLNodeWithVisited(node.Content[0], visited)
			list := parseYAMLNodeWithVisited(node.Content[1], visited)
			return &Intrinsic{Type: IntrinsicSelect, Args: []any{index, list}}
		}

	case "GetAZs":
		if node.Kind == yaml.ScalarNode {
			return &Intrinsic{Type: IntrinsicGetAZs, Args: node.Value}
		}
		if node.Kind == yaml.SequenceNode && len(node.Content) > 0 {
			return &Intrinsic{Type: IntrinsicGetAZs, Args: parseScalarString(node.Content[0])}
		}
		if node.Kind == yaml.MappingNode {
			// Handle nested intrinsic - use parseNodeContentsWithVisited to avoid infinite recursion
			return &Intrinsic{Type: IntrinsicGetAZs, Args: parseNodeContentsWithVisited(node, visited)}
		}
		return &Intrinsic{Type: IntrinsicGetAZs, Args: ""}

	case "If":
		if node.Kind == yaml.SequenceNode && len(node.Content) >= 3 {
			args := make([]any, 3)
			args[0] = parseScalarString(node.Content[0])
			args[1] = parseYAMLNodeWithVisited(node.Content[1], visited)
			args[2] = parseYAMLNodeWithVisited(node.Content[2], visited)
			return &Intrinsic{Type: IntrinsicIf, Args: args}
		}

	case "Equals":
		if node.Kind == yaml.SequenceNode && len(node.Content) >= 2 {
			args := make([]any, 2)
			args[0] = parseYAMLNodeWithVisited(node.Content[0], visited)
			args[1] = parseYAMLNodeWithVisited(node.Content[1], visited)
			return &Intrinsic{Type: IntrinsicEquals, Args: args}
		}

	case "And":
		if node.Kind == yaml.SequenceNode {
			args := make([]any, 0, len(node.Content))
			for _, child := range node.Content {
				args = append(args, parseYAMLNodeWithVisited(child, visited))
			}
			return &Intrinsic{Type: IntrinsicAnd, Args: args}
		}

	case "Or":
		if node.Kind == yaml.SequenceNode {
			args := make([]any, 0, len(node.Content))
			for _, child := range node.Content {
				args = append(args, parseYAMLNodeWithVisited(child, visited))
			}
			return &Intrinsic{Type: IntrinsicOr, Args: args}
		}

	case "Not":
		if node.Kind == yaml.SequenceNode && len(node.Content) > 0 {
			return &Intrinsic{Type: IntrinsicNot, Args: parseYAMLNodeWithVisited(node.Content[0], visited)}
		}

	case "Condition":
		return &Intrinsic{Type: IntrinsicCondition, Args: parseScalarString(node)}

	case "FindInMap":
		if node.Kind == yaml.SequenceNode && len(node.Content) >= 3 {
			args := make([]any, 3)
			for i := 0; i < 3; i++ {
				args[i] = parseYAMLNodeWithVisited(node.Content[i], visited)
			}
			return &Intrinsic{Type: IntrinsicFindInMap, Args: args}
		}

	case "Base64":
		if node.Kind == yaml.ScalarNode {
			return &Intrinsic{Type: IntrinsicBase64, Args: node.Value}
		}
		// For non-scalar (e.g., mapping with Fn::Join), parse contents directly
		return &Intrinsic{Type: IntrinsicBase64, Args: parseNodeContentsWithVisited(node, visited)}

	case "Cidr":
		if node.Kind == yaml.SequenceNode && len(node.Content) >= 3 {
			args := make([]any, 3)
			for i := 0; i < 3; i++ {
				args[i] = parseYAMLNodeWithVisited(node.Content[i], visited)
			}
			return &Intrinsic{Type: IntrinsicCidr, Args: args}
		}

	case "ImportValue":
		if node.Kind == yaml.ScalarNode {
			return &Intrinsic{Type: IntrinsicImportValue, Args: node.Value}
		}
		// For non-scalar (e.g., nested intrinsics), parse contents directly
		return &Intrinsic{Type: IntrinsicImportValue, Args: parseNodeContentsWithVisited(node, visited)}

	case "Split":
		if node.Kind == yaml.SequenceNode && len(node.Content) >= 2 {
			args := make([]any, 2)
			args[0] = parseScalarString(node.Content[0])
			args[1] = parseYAMLNodeWithVisited(node.Content[1], visited)
			return &Intrinsic{Type: IntrinsicSplit, Args: args}
		}

	case "Transform":
		return &Intrinsic{Type: IntrinsicTransform, Args: parseNodeContentsWithVisited(node, visited)}

	case "ValueOf":
		if node.Kind == yaml.SequenceNode && len(node.Content) >= 2 {
			args := make([]any, len(node.Content))
			for i, child := range node.Content {
				args[i] = parseYAMLNodeWithVisited(child, visited)
			}
			return &Intrinsic{Type: IntrinsicValueOf, Args: args}
		}
	}

	// Unknown tag - return the node's value directly without recursion
	if node.Kind == yaml.ScalarNode {
		return &Intrinsic{Type: IntrinsicRef, Args: node.Value}
	}
	// For complex unknown tags, just return nil
	return nil
}

// parseFromMap builds a Template from a parsed map.
func parseFromMap(data map[string]any, sourceName string) (*Template, error) {
	tmpl := NewTemplate()
	tmpl.SourceFile = sourceName

	if desc, ok := data["Description"].(string); ok {
		tmpl.Description = desc
	}
	if ver, ok := data["AWSTemplateFormatVersion"].(string); ok {
		tmpl.AWSTemplateFormatVersion = ver
	}

	// Parse parameters
	if params, ok := data["Parameters"].(map[string]any); ok {
		for logicalID, paramDef := range params {
			if paramMap, ok := paramDef.(map[string]any); ok {
				tmpl.Parameters[logicalID] = parseParameter(logicalID, paramMap)
			}
		}
	}

	// Parse mappings
	if mappings, ok := data["Mappings"].(map[string]any); ok {
		for logicalID, mapData := range mappings {
			if mapMap, ok := mapData.(map[string]any); ok {
				tmpl.Mappings[logicalID] = parseMapping(logicalID, mapMap)
			}
		}
	}

	// Parse conditions
	if conditions, ok := data["Conditions"].(map[string]any); ok {
		for logicalID, expr := range conditions {
			tmpl.Conditions[logicalID] = parseConditionDef(logicalID, expr)
		}
	}

	// Parse resources
	if resources, ok := data["Resources"].(map[string]any); ok {
		for logicalID, resourceDef := range resources {
			// Skip Fn::ForEach meta-resources
			if strings.HasPrefix(logicalID, "Fn::ForEach::") {
				continue
			}
			if resourceMap, ok := resourceDef.(map[string]any); ok {
				tmpl.Resources[logicalID] = parseResource(logicalID, resourceMap)
			}
		}
	}

	// Parse outputs
	if outputs, ok := data["Outputs"].(map[string]any); ok {
		for logicalID, outputDef := range outputs {
			// Skip Fn::ForEach meta-outputs
			if strings.HasPrefix(logicalID, "Fn::ForEach::") {
				continue
			}
			if outputMap, ok := outputDef.(map[string]any); ok {
				tmpl.Outputs[logicalID] = parseOutput(logicalID, outputMap)
			}
		}
	}

	// Build reference graph
	analyzeReferences(tmpl)

	return tmpl, nil
}

func parseParameter(logicalID string, props map[string]any) *Parameter {
	param := &Parameter{
		LogicalID: logicalID,
		Type:      "String",
	}

	if t, ok := props["Type"].(string); ok {
		param.Type = t
	}
	if desc, ok := props["Description"].(string); ok {
		param.Description = desc
	}
	if def, ok := props["Default"]; ok {
		param.Default = def
	}
	if allowed, ok := props["AllowedValues"].([]any); ok {
		param.AllowedValues = allowed
	}
	if pattern, ok := props["AllowedPattern"].(string); ok {
		param.AllowedPattern = pattern
	}
	if minLen, ok := props["MinLength"].(int); ok {
		param.MinLength = &minLen
	}
	if maxLen, ok := props["MaxLength"].(int); ok {
		param.MaxLength = &maxLen
	}
	if minVal, ok := props["MinValue"].(float64); ok {
		param.MinValue = &minVal
	}
	if maxVal, ok := props["MaxValue"].(float64); ok {
		param.MaxValue = &maxVal
	}
	if constraint, ok := props["ConstraintDescription"].(string); ok {
		param.ConstraintDescription = constraint
	}
	if noEcho, ok := props["NoEcho"].(bool); ok {
		param.NoEcho = noEcho
	}

	return param
}

func parseMapping(logicalID string, mapData map[string]any) *Mapping {
	mapping := &Mapping{
		LogicalID: logicalID,
		MapData:   make(map[string]map[string]any),
	}

	for topKey, topVal := range mapData {
		if secondLevel, ok := topVal.(map[string]any); ok {
			mapping.MapData[topKey] = secondLevel
		}
	}

	return mapping
}

func parseConditionDef(logicalID string, expr any) *Condition {
	return &Condition{
		LogicalID:  logicalID,
		Expression: resolveLongFormIntrinsics(expr),
	}
}

func parseResource(logicalID string, resourceDef map[string]any) *Resource {
	resource := &Resource{
		LogicalID:  logicalID,
		Properties: make(map[string]*Property),
	}

	if rt, ok := resourceDef["Type"].(string); ok {
		resource.ResourceType = rt
	}

	if props, ok := resourceDef["Properties"].(map[string]any); ok {
		for cfName, value := range props {
			resource.Properties[cfName] = parseProperty(cfName, value)
		}
	}

	if dependsOn, ok := resourceDef["DependsOn"]; ok {
		switch v := dependsOn.(type) {
		case string:
			resource.DependsOn = []string{v}
		case []any:
			for _, d := range v {
				if s, ok := d.(string); ok {
					resource.DependsOn = append(resource.DependsOn, s)
				}
			}
		}
	}

	if cond, ok := resourceDef["Condition"].(string); ok {
		resource.Condition = cond
	}
	if dp, ok := resourceDef["DeletionPolicy"].(string); ok {
		resource.DeletionPolicy = dp
	}
	if urp, ok := resourceDef["UpdateReplacePolicy"].(string); ok {
		resource.UpdateReplacePolicy = urp
	}
	if metadata, ok := resourceDef["Metadata"].(map[string]any); ok {
		resource.Metadata = metadata
	}

	return resource
}

func parseProperty(cfName string, value any) *Property {
	return &Property{
		Name:  cfName,
		Value: resolveLongFormIntrinsics(value),
	}
}

func parseOutput(logicalID string, outputDef map[string]any) *Output {
	output := &Output{
		LogicalID: logicalID,
	}

	if val, ok := outputDef["Value"]; ok {
		output.Value = resolveLongFormIntrinsics(val)
	}
	if desc, ok := outputDef["Description"].(string); ok {
		output.Description = desc
	}
	if export, ok := outputDef["Export"].(map[string]any); ok {
		if name, ok := export["Name"]; ok {
			output.ExportName = resolveLongFormIntrinsics(name)
		}
	}
	if cond, ok := outputDef["Condition"].(string); ok {
		output.Condition = cond
	}

	return output
}

// resolveLongFormIntrinsics converts JSON-style Fn:: intrinsics to Intrinsic objects.
func resolveLongFormIntrinsics(value any) any {
	if value == nil {
		return nil
	}

	// Already converted
	if _, ok := value.(*Intrinsic); ok {
		return value
	}

	switch v := value.(type) {
	case map[string]any:
		if len(v) == 1 {
			for key, val := range v {
				// Check for Ref
				if key == "Ref" {
					if s, ok := val.(string); ok {
						return &Intrinsic{Type: IntrinsicRef, Args: s}
					}
				}

				// Check for Fn:: prefix
				if strings.HasPrefix(key, "Fn::") {
					intrinsicName := key[4:]
					resolvedVal := resolveLongFormIntrinsics(val)

					switch intrinsicName {
					case "GetAtt":
						switch rv := resolvedVal.(type) {
						case string:
							parts := strings.SplitN(rv, ".", 2)
							return &Intrinsic{Type: IntrinsicGetAtt, Args: parts}
						case []any:
							strs := make([]string, len(rv))
							for i, p := range rv {
								strs[i] = fmt.Sprintf("%v", p)
							}
							return &Intrinsic{Type: IntrinsicGetAtt, Args: strs}
						}

					case "Sub":
						switch rv := resolvedVal.(type) {
						case string:
							return &Intrinsic{Type: IntrinsicSub, Args: rv}
						case []any:
							if len(rv) > 1 {
								return &Intrinsic{Type: IntrinsicSub, Args: rv}
							} else if len(rv) == 1 {
								return &Intrinsic{Type: IntrinsicSub, Args: rv[0]}
							}
						}

					case "Join":
						if arr, ok := resolvedVal.([]any); ok && len(arr) >= 2 {
							return &Intrinsic{Type: IntrinsicJoin, Args: arr}
						}

					case "Select":
						if arr, ok := resolvedVal.([]any); ok && len(arr) >= 2 {
							return &Intrinsic{Type: IntrinsicSelect, Args: arr}
						}

					case "GetAZs":
						if s, ok := resolvedVal.(string); ok {
							return &Intrinsic{Type: IntrinsicGetAZs, Args: s}
						}
						return &Intrinsic{Type: IntrinsicGetAZs, Args: ""}

					case "If":
						if arr, ok := resolvedVal.([]any); ok && len(arr) >= 3 {
							return &Intrinsic{Type: IntrinsicIf, Args: arr[:3]}
						}

					case "Equals":
						if arr, ok := resolvedVal.([]any); ok && len(arr) >= 2 {
							return &Intrinsic{Type: IntrinsicEquals, Args: arr[:2]}
						}

					case "And":
						if arr, ok := resolvedVal.([]any); ok {
							return &Intrinsic{Type: IntrinsicAnd, Args: arr}
						}

					case "Or":
						if arr, ok := resolvedVal.([]any); ok {
							return &Intrinsic{Type: IntrinsicOr, Args: arr}
						}

					case "Not":
						if arr, ok := resolvedVal.([]any); ok && len(arr) > 0 {
							return &Intrinsic{Type: IntrinsicNot, Args: arr[0]}
						}
						return &Intrinsic{Type: IntrinsicNot, Args: resolvedVal}

					case "FindInMap":
						if arr, ok := resolvedVal.([]any); ok && len(arr) >= 3 {
							return &Intrinsic{Type: IntrinsicFindInMap, Args: arr[:3]}
						}

					case "Base64":
						return &Intrinsic{Type: IntrinsicBase64, Args: resolvedVal}

					case "Cidr":
						if arr, ok := resolvedVal.([]any); ok && len(arr) >= 3 {
							return &Intrinsic{Type: IntrinsicCidr, Args: arr[:3]}
						}

					case "ImportValue":
						return &Intrinsic{Type: IntrinsicImportValue, Args: resolvedVal}

					case "Split":
						if arr, ok := resolvedVal.([]any); ok && len(arr) >= 2 {
							return &Intrinsic{Type: IntrinsicSplit, Args: arr[:2]}
						}

					case "Transform":
						return &Intrinsic{Type: IntrinsicTransform, Args: resolvedVal}
					}
				}

				// Check for Condition
				if key == "Condition" {
					if s, ok := val.(string); ok {
						return &Intrinsic{Type: IntrinsicCondition, Args: s}
					}
				}
			}
		}

		// Regular dict - recurse
		result := make(map[string]any, len(v))
		for k, val := range v {
			result[k] = resolveLongFormIntrinsics(val)
		}
		return result

	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = resolveLongFormIntrinsics(item)
		}
		return result
	}

	return value
}

// analyzeReferences builds the reference graph by analyzing Ref and GetAtt usage.
func analyzeReferences(tmpl *Template) {
	subVarPattern := regexp.MustCompile(`\$\{([^}]+)\}`)

	var findRefs func(value any, sourceID string)
	findRefs = func(value any, sourceID string) {
		switch v := value.(type) {
		case *Intrinsic:
			switch v.Type {
			case IntrinsicRef:
				if targetID, ok := v.Args.(string); ok {
					tmpl.ReferenceGraph[sourceID] = append(tmpl.ReferenceGraph[sourceID], targetID)
				}
			case IntrinsicGetAtt:
				if parts, ok := v.Args.([]string); ok && len(parts) > 0 {
					tmpl.ReferenceGraph[sourceID] = append(tmpl.ReferenceGraph[sourceID], parts[0])
				}
			case IntrinsicSub:
				var subStr string
				switch args := v.Args.(type) {
				case string:
					subStr = args
				case []any:
					if len(args) > 0 {
						if s, ok := args[0].(string); ok {
							subStr = s
						}
					}
				}
				// Find ${Var} patterns
				matches := subVarPattern.FindAllStringSubmatch(subStr, -1)
				for _, match := range matches {
					refName := strings.Split(match[1], ".")[0]
					if !strings.HasPrefix(refName, "AWS::") {
						tmpl.ReferenceGraph[sourceID] = append(tmpl.ReferenceGraph[sourceID], refName)
					}
				}
			}

			// Recurse into args
			switch args := v.Args.(type) {
			case *Intrinsic:
				findRefs(args, sourceID)
			case []any:
				for _, item := range args {
					findRefs(item, sourceID)
				}
			case map[string]any:
				for _, item := range args {
					findRefs(item, sourceID)
				}
			}

		case map[string]any:
			for _, item := range v {
				findRefs(item, sourceID)
			}

		case []any:
			for _, item := range v {
				findRefs(item, sourceID)
			}
		}
	}

	for resourceID, resource := range tmpl.Resources {
		for _, prop := range resource.Properties {
			findRefs(prop.Value, resourceID)
		}
	}

	for outputID, output := range tmpl.Outputs {
		findRefs(output.Value, outputID)
	}
}
