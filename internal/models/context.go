package models

// Context represents a Spacelift context.
type Context struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Space       string          `json:"space"`
	Labels      []string        `json:"labels"`
	Hooks       Hooks           `json:"hooks"`
	Config      []ConfigElement `json:"config"`
	CreatedAt   int64           `json:"createdAt"`
	UpdatedAt   int64           `json:"updatedAt"`
}

// ConfigElement represents a configuration element (env var or mounted file).
type ConfigElement struct {
	ID          string `json:"id"`
	Type        string `json:"type"`        // ENVIRONMENT_VARIABLE or FILE_MOUNT
	Value       string `json:"value"`       // Empty for secrets (write-only)
	WriteOnly   bool   `json:"writeOnly"`   // True if this is a secret
	Description string `json:"description"`
}

// IsSecret returns true if this config element is a secret (not readable).
func (c *ConfigElement) IsSecret() bool {
	return c.WriteOnly
}

// GetNonSecretConfigs returns only non-secret config elements from a context.
func (c *Context) GetNonSecretConfigs() []ConfigElement {
	var result []ConfigElement
	for _, elem := range c.Config {
		if !elem.WriteOnly {
			result = append(result, elem)
		}
	}
	return result
}

// GetSecretConfigs returns only secret config elements from a context.
func (c *Context) GetSecretConfigs() []ConfigElement {
	var result []ConfigElement
	for _, elem := range c.Config {
		if elem.WriteOnly {
			result = append(result, elem)
		}
	}
	return result
}

// HasSecrets returns true if the context contains any secret config elements.
func (c *Context) HasSecrets() bool {
	for _, elem := range c.Config {
		if elem.WriteOnly {
			return true
		}
	}
	return false
}
