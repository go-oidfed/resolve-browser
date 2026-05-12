package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-oidfed/lib"
	"github.com/go-oidfed/lib/unixtime"
	"github.com/gofiber/fiber/v2"

	"github.com/go-oidfed/resolve-browser/internal/api"
)

func ResolvePreviewHandler(c *fiber.Ctx) error {
	var req api.PreviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			api.ErrorResponse{
				Error:            "invalid_request",
				ErrorDescription: "failed to parse request body: " + err.Error(),
			},
		)
	}

	if len(req) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			api.ErrorResponse{
				Error:            "invalid_request",
				ErrorDescription: "trust_chain cannot be empty",
			},
		)
	}

	statements := make([]*oidfed.EntityStatement, len(req))
	for i, ch := range req {
		stmt, err := parseEntityStatementInput(ch)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				api.ErrorResponse{
					Error:            "invalid_statement",
					ErrorDescription: fmt.Sprintf("failed to parse statement at index %d: %s", len(req)-i, err.Error()),
				},
			)
		}
		statements[i] = &oidfed.EntityStatement{
			EntityStatementPayload: *stmt,
		}
	}

	for i := len(statements) - 1; i > 0; i-- {
		parent := statements[i]
		child := statements[i-1]

		if parent.Constraints != nil {
			if !checkConstraints(parent.Constraints, child, len(statements)-i) {
				return c.Status(fiber.StatusBadRequest).JSON(
					api.ErrorResponse{
						Error:            "constraint_violation",
						ErrorDescription: fmt.Sprintf("constraint violation at position %d", len(statements)-i),
					},
				)
			}
		}
	}

	leafMetadata := statements[0].Metadata

	metadataPolicies := make([]*oidfed.MetadataPolicies, len(statements))
	for i, stmt := range statements {
		metadataPolicies[i] = stmt.MetadataPolicy
	}

	combinedPolicy, err := oidfed.MergeMetadataPolicies(metadataPolicies...)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			api.ErrorResponse{
				Error:            "invalid_metadata_policy",
				ErrorDescription: "failed to merge metadata policies: " + err.Error(),
			},
		)
	}

	var baseMetadata *oidfed.Metadata
	if len(statements) > 1 && statements[1].Metadata != nil {
		superiorMetadata := statements[1].Metadata
		if leafMetadata == nil {
			baseMetadata = superiorMetadata
		} else {
			baseMetadata = leafMetadata
			mergeMetadata(baseMetadata, superiorMetadata)
		}
	} else if leafMetadata != nil {
		baseMetadata = leafMetadata
	} else {
		baseMetadata = &oidfed.Metadata{}
	}

	finalMetadata, err := baseMetadata.ApplyPolicy(combinedPolicy)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			api.ErrorResponse{
				Error:            "invalid_metadata_policy",
				ErrorDescription: "failed to apply metadata policy: " + err.Error(),
			},
		)
	}

	metadataMap := make(map[string]any)
	if metadataBytes, err := json.Marshal(finalMetadata); err == nil {
		json.Unmarshal(metadataBytes, &metadataMap)
	}

	return c.JSON(
		api.PreviewResponse{
			Metadata: metadataMap,
			Valid:    true,
		},
	)
}

func parseEntityStatementInput(data json.RawMessage) (*oidfed.EntityStatementPayload, error) {
	if isJWT(data) {
		return parseJWT(data)
	}

	if isBase64Encoded(data) {
		return parseBase64JSON(data)
	}

	return parseJSON(data)
}

func isJWT(data json.RawMessage) bool {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return false
	}
	parts := strings.Split(s, ".")
	return len(parts) == 3
}

func parseJWT(data json.RawMessage) (*oidfed.EntityStatementPayload, error) {
	var jwtStr string
	if err := json.Unmarshal(data, &jwtStr); err != nil {
		return nil, err
	}

	parts := strings.Split(jwtStr, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var payload oidfed.EntityStatementPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse JWT payload: %w", err)
	}

	normalizeTimestamps(&payload)
	return &payload, nil
}

func isBase64Encoded(data json.RawMessage) bool {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return false
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return false
	}

	return len(decoded) > 0 && decoded[0] == '{'
}

func parseBase64JSON(data json.RawMessage) (*oidfed.EntityStatementPayload, error) {
	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	var payload oidfed.EntityStatementPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse decoded JSON: %w", err)
	}

	normalizeTimestamps(&payload)
	return &payload, nil
}

func parseJSON(data json.RawMessage) (*oidfed.EntityStatementPayload, error) {
	var payload oidfed.EntityStatementPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	normalizeTimestamps(&payload)
	return &payload, nil
}

func normalizeTimestamps(payload *oidfed.EntityStatementPayload) {
	if payload.IssuedAt.IsZero() {
		payload.IssuedAt = unixtime.Unixtime{Time: time.Now()}
	}
	if payload.ExpiresAt.IsZero() {
		payload.ExpiresAt = unixtime.Unixtime{Time: time.Now().Add(24 * time.Hour)}
	}
}

func checkConstraints(constraints *oidfed.ConstraintSpecification, stmt *oidfed.EntityStatement, depth int) bool {
	if constraints == nil {
		return true
	}

	if constraints.MaxPathLength != nil && *constraints.MaxPathLength < depth {
		return false
	}

	if constraints.AllowedEntityTypes != nil && stmt.Metadata != nil {
		entityTypes := stmt.Metadata.GuessEntityTypes()
		for _, et := range entityTypes {
			if et != "federation_entity" && !contains(constraints.AllowedEntityTypes, et) {
				return false
			}
		}
	}

	if naming := constraints.NamingConstraints; naming != nil {
		for _, excluded := range naming.Excluded {
			if matchesNamingConstraint(excluded, stmt.Subject) {
				return false
			}
		}
		if naming.Permitted != nil {
			permitted := false
			for _, p := range naming.Permitted {
				if matchesNamingConstraint(p, stmt.Subject) {
					permitted = true
					break
				}
			}
			if !permitted {
				return false
			}
		}
	}

	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func matchesNamingConstraint(constraint, entityID string) bool {
	if len(entityID) == 0 {
		return false
	}
	if len(constraint) > 0 && constraint[0] == '.' {
		return len(entityID) >= len(constraint) && entityID[len(entityID)-len(constraint):] == constraint
	}
	return constraint == entityID
}

func mergeMetadata(target, source *oidfed.Metadata) {
	if source == nil {
		return
	}

	targetVal := reflect.ValueOf(target).Elem()
	sourceVal := reflect.ValueOf(source).Elem()
	typ := targetVal.Type()

	for i := 0; i < targetVal.NumField(); i++ {
		fieldName := typ.Field(i).Name

		if fieldName == "Extra" {
			continue
		}

		targetField := targetVal.Field(i)
		sourceField := sourceVal.Field(i)

		if sourceField.Kind() == reflect.Ptr && !sourceField.IsNil() {
			if targetField.IsNil() {
				targetField.Set(sourceField)
			} else {
				mergeStructFields(targetField.Elem(), sourceField.Elem())
			}
		}
	}

	if source.Extra != nil {
		if target.Extra == nil {
			target.Extra = make(map[string]any)
		}
		for k, v := range source.Extra {
			target.Extra[k] = v
		}
	}
}

func mergeStructFields(target, source reflect.Value) {
	st := source.Type()
	for i := 0; i < source.NumField(); i++ {
		f := st.Field(i)
		name := f.Name
		sf := source.Field(i)
		tf := target.FieldByName(name)
		if !tf.IsValid() || !tf.CanSet() {
			continue
		}
		if sf.IsZero() {
			continue
		}
		if sf.Kind() == reflect.Map && tf.Kind() == reflect.Map {
			if tf.IsNil() {
				tf.Set(reflect.MakeMap(tf.Type()))
			}
			for _, k := range sf.MapKeys() {
				tf.SetMapIndex(k, sf.MapIndex(k))
			}
		} else {
			tf.Set(sf)
		}
	}
}
