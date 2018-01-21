package utils

// FieldMapping represents a processing mapping for a field
type FieldMapping struct {
	TypeName string `json:"type"`
	Analyzer string `json:"analyzer"`
}

// GetFieldMapping extracts field mapping from type mapping object
func GetFieldMapping(mapping map[string]interface{}, fieldName string) (*FieldMapping, bool) {
	if propertiesRaw, ok := mapping["properties"]; ok {
		properties := propertiesRaw.(map[string]interface{})
		if config, ok := properties[fieldName]; ok {
			configMap := config.(map[string]interface{})
			var fieldMapping FieldMapping
			fieldMapping.TypeName, _ = configMap["type"].(string)
			fieldMapping.Analyzer, _ = configMap["analyzer"].(string)
			return &fieldMapping, true
		}
		return nil, false
	}
	return nil, false
}
