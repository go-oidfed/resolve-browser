package handlers

import (
	"encoding/json"
	"time"

	"github.com/go-oidfed/lib"
	"github.com/go-oidfed/lib/unixtime"
	"github.com/go-oidfed/resolve-browser/internal/api"
	"github.com/gofiber/fiber/v2"
)

func ResolvePreviewHandler(c *fiber.Ctx) error {
	var req api.PreviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error:            "invalid_request",
			ErrorDescription: "failed to parse request body: " + err.Error(),
		})
	}

	if len(req.TrustChain) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error:            "invalid_request",
			ErrorDescription: "trust_chain cannot be empty",
		})
	}

	if req.TrustAnchor == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error:            "invalid_request",
			ErrorDescription: "trust_anchor is required",
		})
	}

	statements := make([]*oidfed.EntityStatement, len(req.TrustChain))
	for i, editable := range req.TrustChain {
		stmt, err := editableStatementToEntityStatement(editable)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
				Error:            "invalid_statement",
				ErrorDescription: "failed to convert statement at index " + string(rune(i+1)) + ": " + err.Error(),
			})
		}
		statements[i] = stmt
	}

	for i := 0; i < len(statements)-1; i++ {
		parent := statements[i]
		child := statements[i+1]

		if parent.Constraints != nil {
			if !checkConstraints(parent.Constraints, child, i+1) {
				return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
					Error:            "constraint_violation",
					ErrorDescription: "constraint violation at position " + string(rune(i+1)),
				})
			}
		}
	}

	leafIndex := len(statements) - 1
	leafMetadata := statements[leafIndex].Metadata
	if leafMetadata == nil {
		leafMetadata = &oidfed.Metadata{}
	}

	metadataPolicies := make([]*oidfed.MetadataPolicies, len(statements))
	for i, stmt := range statements {
		metadataPolicies[i] = stmt.MetadataPolicy
	}

	combinedPolicy, err := oidfed.MergeMetadataPolicies(metadataPolicies...)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error:            "invalid_metadata_policy",
			ErrorDescription: "failed to merge metadata policies: " + err.Error(),
		})
	}

	finalMetadata, err := leafMetadata.ApplyPolicy(combinedPolicy)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error:            "invalid_metadata_policy",
			ErrorDescription: "failed to apply metadata policy: " + err.Error(),
		})
	}

	metadataMap := make(map[string]any)
	if metadataBytes, err := json.Marshal(finalMetadata); err == nil {
		json.Unmarshal(metadataBytes, &metadataMap)
	}

	return c.JSON(api.PreviewResponse{
		Metadata: metadataMap,
		Valid:    true,
	})
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
