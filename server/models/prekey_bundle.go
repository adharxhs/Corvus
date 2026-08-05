package models

type PrekeyBundle struct {
	UserID                string
	IdentityKey           []byte
	SignedPrekey          []byte
	SignedPrekeySignature []byte
	OneTimePrekey         []byte
}

type PrekeyBundleRequest struct {
	IdentityKey           string `json:"identity_key"`
	SignedPrekey          string `json:"signed_prekey"`
	SignedPrekeySignature string `json:"signed_prekey_signature"`
	OneTimePrekey         string `json:"one_time_prekey,omitempty"`
}

type PrekeyBundleResponse struct {
	UserID                string `json:"user_id"`
	IdentityKey           string `json:"identity_key"`
	SignedPrekey          string `json:"signed_prekey"`
	SignedPrekeySignature string `json:"signed_prekey_signature"`
	OneTimePrekey         string `json:"one_time_prekey,omitempty"`
}
