package schema

// ABDMetadata is a struct representing aa JSON blob in the ABD Metadata
// Format, as defined by the ABD spec.
type ABDMetadata struct {
	Identifier string            `json:"identifier"`
	Labels     map[string]string `json:"labels"`
	Mirrors    []ABDMirror       `json:"mirrors"`
}

// ABDMirror is a struct representing a mirror from which an artifact can be
// fetched.
type ABDMirror struct {
	Artifact  string `json:"artifact"`
	Signature string `json:"signature"`
}

// ABDError is a struct representing a JSON blob detailing an error returned by
// a strategy.
type ABDError struct {
	E string `json:"error"`
}

func (e *ABDError) Error() string {
	return e.E
}

// ABDMetadataFetchStrategyConfiguration is a struct representing a JSON blob
// of a configuration file for ABD.
type ABDMetadataFetchStrategyConfiguration struct {
	Prefix   string `json:"prefix"`
	Strategy string `json:"strategy"`
}
