package api

import "encoding/json"

type ResolveRequest struct {
	Subject     string   `json:"sub"`
	TrustAnchor []string `json:"trust_anchor"`
	EntityTypes []string `json:"entity_types,omitempty"`
}

type ResolveResponse struct {
	Metadata   map[string]any  `json:"metadata"`
	TrustChain []StatementInfo `json:"trust_chain"`
	TrustMarks []TrustMarkInfo `json:"trust_marks,omitempty"`
	ExpiresAt  int64           `json:"expires_at"`
}

type StatementInfo struct {
	Index      int            `json:"index"`
	Type       string         `json:"type"`
	Issuer     string         `json:"issuer"`
	Subject    string         `json:"subject"`
	JWTHdr     map[string]any `json:"jwt_header"`
	JWTPayload map[string]any `json:"jwt_payload"`
	RawJWT     string         `json:"raw_jwt"`
	IssuedAt   int64          `json:"iat,omitempty"`
	ExpiresAt  int64          `json:"exp,omitempty"`
}

type TrustMarkInfo struct {
	TrustMarkType string `json:"trust_mark_type"`
	TrustMark     string `json:"trust_mark"`
}

type PreviewRequest []json.RawMessage

type ConstraintSpec struct {
	MaxPathLength      *int               `json:"max_path_length,omitempty"`
	NamingConstraints  *NamingConstraints `json:"naming_constraints,omitempty"`
	AllowedEntityTypes []string           `json:"allowed_entity_types,omitempty"`
}

type NamingConstraints struct {
	Permitted []string `json:"permitted,omitempty"`
	Excluded  []string `json:"excluded,omitempty"`
}

type PreviewResponse struct {
	Metadata map[string]any `json:"metadata"`
	Valid    bool           `json:"valid"`
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
