package handlers

import (
	"encoding/base64"
	"encoding/json"

	"github.com/go-oidfed/lib"
	"github.com/go-oidfed/resolve-browser/internal/api"
	"github.com/gofiber/fiber/v2"
)

func ResolveHandler(c *fiber.Ctx) error {
	var req api.ResolveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error:            "invalid_request",
			ErrorDescription: "failed to parse request body: " + err.Error(),
		})
	}

	if req.Subject == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error:            "invalid_request",
			ErrorDescription: "required parameter 'sub' not given",
		})
	}

	if len(req.TrustAnchor) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error:            "invalid_request",
			ErrorDescription: "required parameter 'trust_anchor' not given",
		})
	}

	resolver := oidfed.TrustResolver{
		TrustAnchors:   oidfed.NewTrustAnchorsFromEntityIDs(req.TrustAnchor...),
		StartingEntity: req.Subject,
		Types:          req.EntityTypes,
	}

	chains := resolver.ResolveToValidChains()
	if len(chains) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(api.ErrorResponse{
			Error:            "invalid_trust_chain",
			ErrorDescription: "no valid trust path between sub and anchor found",
		})
	}

	selectedChain := chains.Filter(oidfed.TrustChainsFilterMinPathLength)[0]
	metadata, err := selectedChain.Metadata()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error:            "metadata_error",
			ErrorDescription: "failed to extract metadata: " + err.Error(),
		})
	}

	trustChain := make([]api.StatementInfo, len(selectedChain))
	for i, stmt := range selectedChain {
		reversedIndex := len(selectedChain) - 1 - i
		trustChain[reversedIndex] = extractStatementInfo(stmt, reversedIndex, len(selectedChain))
	}

	var trustMarks []api.TrustMarkInfo
	leaf := selectedChain[0]
	if leaf.TrustMarks != nil {
		ta := selectedChain[len(selectedChain)-1]
		verifiedMarks := leaf.TrustMarks.VerifiedFederation(&ta.EntityStatementPayload)
		for _, mark := range verifiedMarks {
			trustMarks = append(trustMarks, api.TrustMarkInfo{
				TrustMarkType: mark.TrustMarkType,
				TrustMark:     mark.TrustMarkJWT,
			})
		}
	}

	metadataMap := make(map[string]any)
	if metadataBytes, err := json.Marshal(metadata); err == nil {
		json.Unmarshal(metadataBytes, &metadataMap)
	}

	return c.JSON(api.ResolveResponse{
		Metadata:   metadataMap,
		TrustChain: trustChain,
		TrustMarks: trustMarks,
		ExpiresAt:  selectedChain.ExpiresAt().Unix(),
	})
}

func extractStatementInfo(stmt *oidfed.EntityStatement, index int, chainLen int) api.StatementInfo {
	var stmtType string
	
	if index == 0 && stmt.Issuer == stmt.Subject {
		stmtType = "trust_anchor_entity_configuration"
	} else if index == chainLen-1 {
		stmtType = "entity_configuration"
	} else {
		stmtType = "subordinate_statement"
	}

	jwtHdr := make(map[string]any)
	jwtPayload := make(map[string]any)
	rawJWT := ""

	payloadBytes, _ := json.Marshal(stmt.EntityStatementPayload)
	json.Unmarshal(payloadBytes, &jwtPayload)

	type jwtHeader struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
		Kid string `json:"kid,omitempty"`
	}
	hdr := jwtHeader{
		Alg: "RS256",
		Typ: "entity-statement+jwt",
	}
	hdrBytes, _ := json.Marshal(hdr)
	json.Unmarshal(hdrBytes, &jwtHdr)

	rawJWT = createMockJWT(jwtHdr, jwtPayload)

	return api.StatementInfo{
		Index:      index,
		Type:       stmtType,
		Issuer:     stmt.Issuer,
		Subject:    stmt.Subject,
		JWTHdr:     jwtHdr,
		JWTPayload: jwtPayload,
		RawJWT:     rawJWT,
		IssuedAt:   stmt.IssuedAt.Unix(),
		ExpiresAt:  stmt.ExpiresAt.Unix(),
	}
}

func createMockJWT(header, payload map[string]any) string {
	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signatureEncoded := base64.RawURLEncoding.EncodeToString([]byte("mock-signature"))

	return headerEncoded + "." + payloadEncoded + "." + signatureEncoded
}
