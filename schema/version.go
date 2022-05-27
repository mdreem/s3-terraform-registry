package schema

type Platform struct {
	Os   string `json:"os"`
	Arch string `json:"arch"`
}

type ProviderVersion struct {
	Version   string     `json:"version"`
	Protocols []string   `json:"protocols"`
	Platforms []Platform `json:"platforms"`
}

type ProviderVersions struct {
	ID       string            `json:"id"`
	Versions []ProviderVersion `json:"versions"`
	Warnings []string          `json:"warnings"`
}
