package easyjson

// Example: How the screenshot code could be improved

// OLD WAY (from screenshot):
func BuildInventoryPayLoadOld(ipAddress string, discoveryResponse interface{}) *JSON {
	payLoad := NewJSONObject()
	schemas := ConvertDiscoveryToInventorySchemas(discoveryResponse)

	payLoad.SetByPath("ip_address", NewJSON(ipAddress))

	schemasArray := make([]any, len(schemas))
	for i, schema := range schemas {
		schemaMap := map[string]any{
			"type":       schema.Type,
			"attributes": schema.Attributes,
		}

		if len(schema.Links) > 0 {
			linksArray := make([]any, len(schema.Links))
			for i, link := range schema.Links {
				linkMap := map[string]any{
					"from": link.From,
					"to":   link.To,
					"name": link.Name,
				}
				linksArray[i] = linkMap
			}
			schemaMap["links"] = linksArray
		} else {
			schemaMap["links"] = []any{}
		}

		schemasArray[i] = schemaMap
	}

	payLoad.SetByPath("schemas", NewJSON(schemasArray))
	return payLoad.GetPtr()
}

// NEW WAY 1: Using bulk operations
func BuildInventoryPayLoadBulk(ipAddress string, discoveryResponse interface{}) *JSON {
	schemas := ConvertDiscoveryToInventorySchemas(discoveryResponse)

	payLoad := NewJSONObjectFromMap(map[string]interface{}{
		"ip_address": ipAddress,
		"schemas":    convertSchemasToArray(schemas),
	})

	return payLoad.GetPtr()
}

// NEW WAY 2: Using builder pattern
func BuildInventoryPayLoadBuilder(ipAddress string, discoveryResponse interface{}) *JSON {
	schemas := ConvertDiscoveryToInventorySchemas(discoveryResponse)

	builder := NewJSONBuilder().
		Set("ip_address", ipAddress).
		Set("schemas", convertSchemasToArray(schemas))

	payload := builder.Build()
	return payload.GetPtr()
}

// NEW WAY 3: Using generic array builder
func BuildInventoryPayLoadGeneric(ipAddress string, discoveryResponse interface{}) *JSON {
	schemas := ConvertDiscoveryToInventorySchemas(discoveryResponse)

	schemasJSON := BuildArrayFromSlice(schemas, func(schema Schema) map[string]interface{} {
		result := map[string]interface{}{
			"type":       schema.Type,
			"attributes": schema.Attributes,
		}

		if len(schema.Links) > 0 {
			result["links"] = transformLinks(schema.Links)
		} else {
			result["links"] = []interface{}{}
		}

		return result
	})

	payLoad := NewJSONObjectFromMap(map[string]interface{}{
		"ip_address": ipAddress,
		"schemas":    schemasJSON.Value,
	})

	return payLoad.GetPtr()
}

// Helper functions
func convertSchemasToArray(schemas []Schema) []interface{} {
	result := make([]interface{}, len(schemas))
	for i, schema := range schemas {
		schemaMap := map[string]interface{}{
			"type":       schema.Type,
			"attributes": schema.Attributes,
		}

		if len(schema.Links) > 0 {
			schemaMap["links"] = transformLinks(schema.Links)
		} else {
			schemaMap["links"] = []interface{}{}
		}

		result[i] = schemaMap
	}
	return result
}

func transformLinks(links []Link) []interface{} {
	result := make([]interface{}, len(links))
	for i, link := range links {
		result[i] = map[string]interface{}{
			"from": link.From,
			"to":   link.To,
			"name": link.Name,
		}
	}
	return result
}

// Mock types for example (these would be defined elsewhere)
type Schema struct {
	Type       string
	Attributes interface{}
	Links      []Link
}

type Link struct {
	From string
	To   string
	Name string
}

func ConvertDiscoveryToInventorySchemas(discoveryResponse interface{}) []Schema {
	// Mock implementation
	return []Schema{}
}
