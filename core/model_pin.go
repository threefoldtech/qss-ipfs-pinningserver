package tfpin

// Pin - Pin object
type Pin struct {

	// Content Identifier (CID) to be pinned recursively
	Cid string `json:"cid"`

	// Optional name for pinned data; can be used for lookups later
	Name string `json:"name,omitempty"`

	Origins Origins `json:"origins,omitempty"`

	Meta PinMeta `json:"meta,omitempty"`
}
