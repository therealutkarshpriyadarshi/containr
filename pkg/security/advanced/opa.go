package advanced

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// OPAEngine represents an OPA (Open Policy Agent) policy engine
type OPAEngine struct {
	policies map[string]*Policy
	mu       sync.RWMutex
	config   *OPAConfig
	cache    *policyCache
}

// OPAConfig holds OPA engine configuration
type OPAConfig struct {
	// EnableCache enables policy decision caching
	EnableCache bool
	// CacheTTL is the time-to-live for cached decisions
	CacheTTL time.Duration
	// BundlePath is the path to policy bundles
	BundlePath string
	// EnableMetrics enables OPA metrics collection
	EnableMetrics bool
}

// Policy represents an OPA policy
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Rules       []string               `json:"rules"`
	Package     string                 `json:"package"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PolicyInput represents input data for policy evaluation
type PolicyInput struct {
	Resource   string                 `json:"resource"`
	Action     string                 `json:"action"`
	Subject    string                 `json:"subject"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// PolicyDecision represents the result of a policy evaluation
type PolicyDecision struct {
	Allow      bool                   `json:"allow"`
	Deny       bool                   `json:"deny"`
	Reason     string                 `json:"reason,omitempty"`
	PolicyID   string                 `json:"policy_id"`
	Violations []string               `json:"violations,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// policyCache caches policy decisions
type policyCache struct {
	decisions map[string]*cachedDecision
	mu        sync.RWMutex
	ttl       time.Duration
}

type cachedDecision struct {
	decision  *PolicyDecision
	expiresAt time.Time
}

// NewOPAEngine creates a new OPA policy engine
func NewOPAEngine(config *OPAConfig) (*OPAEngine, error) {
	if config == nil {
		config = defaultOPAConfig()
	}

	engine := &OPAEngine{
		policies: make(map[string]*Policy),
		config:   config,
	}

	if config.EnableCache {
		engine.cache = &policyCache{
			decisions: make(map[string]*cachedDecision),
			ttl:       config.CacheTTL,
		}
	}

	return engine, nil
}

// defaultOPAConfig returns default OPA configuration
func defaultOPAConfig() *OPAConfig {
	return &OPAConfig{
		EnableCache:   true,
		CacheTTL:      5 * time.Minute,
		EnableMetrics: true,
	}
}

// AddPolicy adds a policy to the engine
func (e *OPAEngine) AddPolicy(ctx context.Context, policy *Policy) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	if policy.Package == "" {
		policy.Package = "containr.policies"
	}

	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	e.policies[policy.ID] = policy

	// Invalidate cache when policy is added
	if e.cache != nil {
		e.cache.invalidate()
	}

	return nil
}

// RemovePolicy removes a policy from the engine
func (e *OPAEngine) RemovePolicy(ctx context.Context, policyID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.policies[policyID]; !exists {
		return fmt.Errorf("policy %s not found", policyID)
	}

	delete(e.policies, policyID)

	// Invalidate cache when policy is removed
	if e.cache != nil {
		e.cache.invalidate()
	}

	return nil
}

// GetPolicy retrieves a policy by ID
func (e *OPAEngine) GetPolicy(policyID string) (*Policy, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policy, exists := e.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("policy %s not found", policyID)
	}

	return policy, nil
}

// ListPolicies returns all policies
func (e *OPAEngine) ListPolicies() []*Policy {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policies := make([]*Policy, 0, len(e.policies))
	for _, policy := range e.policies {
		policies = append(policies, policy)
	}

	return policies
}

// Evaluate evaluates input against all policies
func (e *OPAEngine) Evaluate(ctx context.Context, input *PolicyInput) (*PolicyDecision, error) {
	// Check cache first
	if e.cache != nil {
		if decision := e.cache.get(input); decision != nil {
			return decision, nil
		}
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	decision := &PolicyDecision{
		Allow:      true, // Default allow unless explicitly denied
		Deny:       false,
		Violations: []string{},
		Metadata:   make(map[string]interface{}),
	}

	// Evaluate each policy
	for _, policy := range e.policies {
		result, err := e.evaluatePolicy(ctx, policy, input)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate policy %s: %w", policy.ID, err)
		}

		// If any policy denies, the overall decision is deny
		if result.Deny {
			decision.Allow = false
			decision.Deny = true
			decision.PolicyID = policy.ID
			decision.Reason = result.Reason
			decision.Violations = append(decision.Violations, result.Violations...)
			break
		}

		// Collect violations even if allowed
		if len(result.Violations) > 0 {
			decision.Violations = append(decision.Violations, result.Violations...)
		}
	}

	// Cache the decision
	if e.cache != nil {
		e.cache.set(input, decision)
	}

	return decision, nil
}

// evaluatePolicy evaluates input against a specific policy
func (e *OPAEngine) evaluatePolicy(ctx context.Context, policy *Policy, input *PolicyInput) (*PolicyDecision, error) {
	decision := &PolicyDecision{
		Allow:      true,
		Deny:       false,
		PolicyID:   policy.ID,
		Violations: []string{},
	}

	// Simulate policy evaluation (in production, this would use OPA's Rego engine)
	// This is a simplified implementation for demonstration

	// Example rule evaluation based on policy rules
	for _, rule := range policy.Rules {
		violation := e.evaluateRule(rule, input)
		if violation != "" {
			decision.Violations = append(decision.Violations, violation)
			decision.Deny = true
			decision.Allow = false
			decision.Reason = fmt.Sprintf("Policy %s violated: %s", policy.Name, violation)
		}
	}

	return decision, nil
}

// evaluateRule evaluates a single rule against input
func (e *OPAEngine) evaluateRule(rule string, input *PolicyInput) string {
	// Simplified rule evaluation
	// In production, this would parse and execute Rego rules

	// Example rules:
	// - "deny_privileged": Deny privileged containers
	// - "require_readonly_rootfs": Require read-only root filesystem
	// - "deny_host_network": Deny host network mode

	switch rule {
	case "deny_privileged":
		if privileged, ok := input.Attributes["privileged"].(bool); ok && privileged {
			return "privileged containers are not allowed"
		}
	case "require_readonly_rootfs":
		if readOnly, ok := input.Attributes["readonly_rootfs"].(bool); ok && !readOnly {
			return "containers must have read-only root filesystem"
		}
	case "deny_host_network":
		if hostNetwork, ok := input.Attributes["host_network"].(bool); ok && hostNetwork {
			return "host network mode is not allowed"
		}
	case "require_resource_limits":
		if limits, ok := input.Attributes["resource_limits"].(map[string]interface{}); !ok || limits == nil {
			return "resource limits are required"
		}
	case "deny_root_user":
		if runAsUser, ok := input.Attributes["run_as_user"].(int); ok && runAsUser == 0 {
			return "running as root user is not allowed"
		}
	}

	return ""
}

// Query executes a custom query against policies
func (e *OPAEngine) Query(ctx context.Context, query string, input map[string]interface{}) (map[string]interface{}, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Simplified query execution
	// In production, this would use OPA's query API

	result := make(map[string]interface{})
	result["query"] = query
	result["input"] = input
	result["result"] = true

	return result, nil
}

// ValidatePolicy validates a policy definition
func (e *OPAEngine) ValidatePolicy(policy *Policy) error {
	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if len(policy.Rules) == 0 {
		return fmt.Errorf("policy must have at least one rule")
	}

	// In production, this would compile and validate Rego rules
	return nil
}

// GetMetrics returns OPA engine metrics
func (e *OPAEngine) GetMetrics() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	metrics := make(map[string]interface{})
	metrics["total_policies"] = len(e.policies)
	metrics["cache_enabled"] = e.cache != nil

	if e.cache != nil {
		e.cache.mu.RLock()
		metrics["cache_size"] = len(e.cache.decisions)
		e.cache.mu.RUnlock()
	}

	return metrics
}

// policyCache methods

func (c *policyCache) get(input *PolicyInput) *PolicyDecision {
	if c == nil {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.cacheKey(input)
	cached, exists := c.decisions[key]
	if !exists {
		return nil
	}

	// Check if expired
	if time.Now().After(cached.expiresAt) {
		return nil
	}

	return cached.decision
}

func (c *policyCache) set(input *PolicyInput, decision *PolicyDecision) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.cacheKey(input)
	c.decisions[key] = &cachedDecision{
		decision:  decision,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *policyCache) invalidate() {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.decisions = make(map[string]*cachedDecision)
}

func (c *policyCache) cacheKey(input *PolicyInput) string {
	// Generate cache key from input
	data, _ := json.Marshal(input)
	return string(data)
}

// DefaultSecurityPolicies returns default security policies
func DefaultSecurityPolicies() []*Policy {
	return []*Policy{
		{
			ID:          "pod-security-baseline",
			Name:        "Pod Security Baseline",
			Description: "Baseline pod security policy following Kubernetes standards",
			Package:     "containr.policies.pod_security",
			Rules: []string{
				"deny_privileged",
				"deny_host_network",
				"require_resource_limits",
			},
		},
		{
			ID:          "pod-security-restricted",
			Name:        "Pod Security Restricted",
			Description: "Restricted pod security policy with enhanced security",
			Package:     "containr.policies.pod_security",
			Rules: []string{
				"deny_privileged",
				"deny_host_network",
				"require_readonly_rootfs",
				"require_resource_limits",
				"deny_root_user",
			},
		},
		{
			ID:          "image-security",
			Name:        "Image Security Policy",
			Description: "Security policy for container images",
			Package:     "containr.policies.image_security",
			Rules: []string{
				"require_signature",
				"deny_latest_tag",
				"require_trusted_registry",
			},
		},
	}
}
