package handlers

import (
	"encoding/json"
	"reflect"
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

	if len(req.TrustChain) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			api.ErrorResponse{
				Error:            "invalid_request",
				ErrorDescription: "trust_chain cannot be empty",
			},
		)
	}

	if req.TrustAnchor == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			api.ErrorResponse{
				Error:            "invalid_request",
				ErrorDescription: "trust_anchor is required",
			},
		)
	}

	statements := make([]*oidfed.EntityStatement, len(req.TrustChain))
	for i, editable := range req.TrustChain {
		stmt, err := editableStatementToEntityStatement(editable)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				api.ErrorResponse{
					Error:            "invalid_statement",
					ErrorDescription: "failed to convert statement at index " + string(rune(i+1)) + ": " + err.Error(),
				},
			)
		}
		statements[i] = stmt
	}

	for i := 0; i < len(statements)-1; i++ {
		parent := statements[i]
		child := statements[i+1]

		if parent.Constraints != nil {
			if !checkConstraints(parent.Constraints, child, i+1) {
				return c.Status(fiber.StatusBadRequest).JSON(
					api.ErrorResponse{
						Error:            "constraint_violation",
						ErrorDescription: "constraint violation at position " + string(rune(i+1)),
					},
				)
			}
		}
	}

	leafIndex := len(statements) - 1
	leafMetadata := statements[leafIndex].Metadata

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
	if len(statements) > 1 && statements[leafIndex-1].Metadata != nil {
		superiorMetadata := statements[leafIndex-1].Metadata
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

func editableStatementToEntityStatement(editable api.EditableStatement) (*oidfed.EntityStatement, error) {
	payload := oidfed.EntityStatementPayload{
		Issuer:           editable.Issuer,
		Subject:          editable.Subject,
		JWKS:             editable.JWKS,
		AuthorityHints:   editable.AuthorityHints,
		TrustAnchorHints: editable.TrustAnchorHints,
	}

	if editable.IssuedAt != 0 {
		payload.IssuedAt = unixtime.Unixtime{Time: time.Unix(editable.IssuedAt, 0)}
	} else {
		payload.IssuedAt = unixtime.Unixtime{Time: time.Now()}
	}

	if editable.ExpiresAt != 0 {
		payload.ExpiresAt = unixtime.Unixtime{Time: time.Unix(editable.ExpiresAt, 0)}
	} else {
		payload.ExpiresAt = unixtime.Unixtime{Time: time.Now().Add(24 * time.Hour)}
	}

	if editable.Metadata != nil {
		metadataBytes, err := json.Marshal(editable.Metadata)
		if err != nil {
			return nil, err
		}
		var metadata oidfed.Metadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, err
		}
		payload.Metadata = &metadata
	}

	if editable.MetadataPolicy != nil {
		policyBytes, err := json.Marshal(editable.MetadataPolicy)
		if err != nil {
			return nil, err
		}

		var metadataPolicy oidfed.MetadataPolicies
		if err := json.Unmarshal(policyBytes, &metadataPolicy); err != nil {
			return nil, err
		}
		payload.MetadataPolicy = &metadataPolicy
	}

	if editable.Constraints != nil {
		payload.Constraints = &oidfed.ConstraintSpecification{
			MaxPathLength:      editable.Constraints.MaxPathLength,
			AllowedEntityTypes: editable.Constraints.AllowedEntityTypes,
		}
		if editable.Constraints.NamingConstraints != nil {
			payload.Constraints.NamingConstraints = &oidfed.NamingConstraints{
				Permitted: editable.Constraints.NamingConstraints.Permitted,
				Excluded:  editable.Constraints.NamingConstraints.Excluded,
			}
		}
	}

	return &oidfed.EntityStatement{
		EntityStatementPayload: payload,
	}, nil
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
