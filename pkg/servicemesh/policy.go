package servicemesh

import (
	"fmt"
	"time"
)

// TrafficPolicy defines traffic management policies for a service
type TrafficPolicy struct {
	Name           string              `json:"name" yaml:"name"`
	Description    string              `json:"description,omitempty" yaml:"description,omitempty"`
	LoadBalancing  *LoadBalancingPolicy `json:"load_balancing,omitempty" yaml:"load_balancing,omitempty"`
	Retry          *RetryPolicy        `json:"retry,omitempty" yaml:"retry,omitempty"`
	CircuitBreaker *CircuitBreaker     `json:"circuit_breaker,omitempty" yaml:"circuit_breaker,omitempty"`
	Timeout        *TimeoutPolicy      `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	RateLimit      *RateLimitPolicy    `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty"`
	HealthCheck    *HealthCheckPolicy  `json:"health_check,omitempty" yaml:"health_check,omitempty"`
	FaultInjection *FaultInjectionPolicy `json:"fault_injection,omitempty" yaml:"fault_injection,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// LoadBalancingPolicy defines load balancing strategy
type LoadBalancingPolicy struct {
	Algorithm          string                    `json:"algorithm" yaml:"algorithm"` // ROUND_ROBIN, LEAST_REQUEST, RING_HASH, RANDOM, WEIGHTED
	ConsistentHashConfig *ConsistentHashConfig   `json:"consistent_hash_config,omitempty" yaml:"consistent_hash_config,omitempty"`
	LocalityWeights    map[string]int            `json:"locality_weights,omitempty" yaml:"locality_weights,omitempty"`
}

// ConsistentHashConfig defines consistent hashing configuration
type ConsistentHashConfig struct {
	HashOn         string `json:"hash_on" yaml:"hash_on"` // header, cookie, source_ip
	HeaderName     string `json:"header_name,omitempty" yaml:"header_name,omitempty"`
	CookieName     string `json:"cookie_name,omitempty" yaml:"cookie_name,omitempty"`
	MinimumRingSize int   `json:"minimum_ring_size,omitempty" yaml:"minimum_ring_size,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	Attempts      int      `json:"attempts" yaml:"attempts"`
	PerTryTimeout string   `json:"per_try_timeout" yaml:"per_try_timeout"`
	RetryOn       []string `json:"retry_on" yaml:"retry_on"` // 5xx, gateway-error, reset, etc.
	RetriableStatusCodes []int `json:"retriable_status_codes,omitempty" yaml:"retriable_status_codes,omitempty"`
	BackoffStrategy *BackoffStrategy `json:"backoff_strategy,omitempty" yaml:"backoff_strategy,omitempty"`
}

// BackoffStrategy defines retry backoff strategy
type BackoffStrategy struct {
	Type           string  `json:"type" yaml:"type"` // exponential, fixed
	BaseInterval   string  `json:"base_interval" yaml:"base_interval"`
	MaxInterval    string  `json:"max_interval,omitempty" yaml:"max_interval,omitempty"`
	Multiplier     float64 `json:"multiplier,omitempty" yaml:"multiplier,omitempty"`
}

// CircuitBreaker defines circuit breaker settings
type CircuitBreaker struct {
	MaxConnections        int     `json:"max_connections" yaml:"max_connections"`
	MaxPendingRequests    int     `json:"max_pending_requests" yaml:"max_pending_requests"`
	MaxRequests           int     `json:"max_requests" yaml:"max_requests"`
	MaxRetries            int     `json:"max_retries" yaml:"max_retries"`
	ErrorThresholdPercent float64 `json:"error_threshold_percent" yaml:"error_threshold_percent"`
	SleepWindow           string  `json:"sleep_window" yaml:"sleep_window"`
	MinimumHosts          int     `json:"minimum_hosts,omitempty" yaml:"minimum_hosts,omitempty"`
	ConsecutiveErrors     int     `json:"consecutive_errors,omitempty" yaml:"consecutive_errors,omitempty"`
}

// TimeoutPolicy defines timeout settings
type TimeoutPolicy struct {
	ConnectionTimeout string `json:"connection_timeout" yaml:"connection_timeout"`
	RequestTimeout    string `json:"request_timeout" yaml:"request_timeout"`
	IdleTimeout       string `json:"idle_timeout,omitempty" yaml:"idle_timeout,omitempty"`
}

// RateLimitPolicy defines rate limiting settings
type RateLimitPolicy struct {
	RequestsPerUnit int    `json:"requests_per_unit" yaml:"requests_per_unit"`
	Unit            string `json:"unit" yaml:"unit"` // second, minute, hour
	BurstSize       int    `json:"burst_size,omitempty" yaml:"burst_size,omitempty"`
	FillRate        int    `json:"fill_rate,omitempty" yaml:"fill_rate,omitempty"`
}

// HealthCheckPolicy defines health check settings
type HealthCheckPolicy struct {
	Protocol           string `json:"protocol" yaml:"protocol"` // HTTP, TCP, gRPC
	Path               string `json:"path,omitempty" yaml:"path,omitempty"`
	Interval           string `json:"interval" yaml:"interval"`
	Timeout            string `json:"timeout" yaml:"timeout"`
	HealthyThreshold   int    `json:"healthy_threshold" yaml:"healthy_threshold"`
	UnhealthyThreshold int    `json:"unhealthy_threshold" yaml:"unhealthy_threshold"`
	Port               int    `json:"port,omitempty" yaml:"port,omitempty"`
}

// FaultInjectionPolicy defines fault injection for testing
type FaultInjectionPolicy struct {
	Delay  *FaultDelay  `json:"delay,omitempty" yaml:"delay,omitempty"`
	Abort  *FaultAbort  `json:"abort,omitempty" yaml:"abort,omitempty"`
}

// FaultDelay defines delay injection
type FaultDelay struct {
	Percentage float64 `json:"percentage" yaml:"percentage"` // 0-100
	FixedDelay string  `json:"fixed_delay" yaml:"fixed_delay"`
}

// FaultAbort defines abort injection
type FaultAbort struct {
	Percentage  float64 `json:"percentage" yaml:"percentage"` // 0-100
	HTTPStatus  int     `json:"http_status" yaml:"http_status"`
}

// NewTrafficPolicy creates a new traffic policy
func NewTrafficPolicy(name string) *TrafficPolicy {
	return &TrafficPolicy{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Validate validates the traffic policy
func (p *TrafficPolicy) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	// Validate load balancing
	if p.LoadBalancing != nil {
		if err := p.LoadBalancing.Validate(); err != nil {
			return fmt.Errorf("invalid load balancing policy: %w", err)
		}
	}

	// Validate retry policy
	if p.Retry != nil {
		if err := p.Retry.Validate(); err != nil {
			return fmt.Errorf("invalid retry policy: %w", err)
		}
	}

	// Validate circuit breaker
	if p.CircuitBreaker != nil {
		if err := p.CircuitBreaker.Validate(); err != nil {
			return fmt.Errorf("invalid circuit breaker: %w", err)
		}
	}

	// Validate timeout policy
	if p.Timeout != nil {
		if err := p.Timeout.Validate(); err != nil {
			return fmt.Errorf("invalid timeout policy: %w", err)
		}
	}

	// Validate rate limit
	if p.RateLimit != nil {
		if err := p.RateLimit.Validate(); err != nil {
			return fmt.Errorf("invalid rate limit policy: %w", err)
		}
	}

	// Validate health check
	if p.HealthCheck != nil {
		if err := p.HealthCheck.Validate(); err != nil {
			return fmt.Errorf("invalid health check policy: %w", err)
		}
	}

	// Validate fault injection
	if p.FaultInjection != nil {
		if err := p.FaultInjection.Validate(); err != nil {
			return fmt.Errorf("invalid fault injection policy: %w", err)
		}
	}

	return nil
}

// Validate validates the load balancing policy
func (lb *LoadBalancingPolicy) Validate() error {
	validAlgorithms := map[string]bool{
		"ROUND_ROBIN":  true,
		"LEAST_REQUEST": true,
		"RING_HASH":    true,
		"RANDOM":       true,
		"WEIGHTED":     true,
	}

	if !validAlgorithms[lb.Algorithm] {
		return fmt.Errorf("invalid load balancing algorithm: %s", lb.Algorithm)
	}

	// Validate consistent hash config if using RING_HASH
	if lb.Algorithm == "RING_HASH" && lb.ConsistentHashConfig != nil {
		validHashOn := map[string]bool{
			"header":    true,
			"cookie":    true,
			"source_ip": true,
		}

		if !validHashOn[lb.ConsistentHashConfig.HashOn] {
			return fmt.Errorf("invalid hash_on value: %s", lb.ConsistentHashConfig.HashOn)
		}
	}

	return nil
}

// Validate validates the retry policy
func (r *RetryPolicy) Validate() error {
	if r.Attempts <= 0 {
		return fmt.Errorf("retry attempts must be positive")
	}

	if r.Attempts > 10 {
		return fmt.Errorf("retry attempts cannot exceed 10")
	}

	if r.PerTryTimeout == "" {
		return fmt.Errorf("per_try_timeout is required")
	}

	if _, err := time.ParseDuration(r.PerTryTimeout); err != nil {
		return fmt.Errorf("invalid per_try_timeout: %w", err)
	}

	if len(r.RetryOn) == 0 {
		return fmt.Errorf("retry_on conditions are required")
	}

	// Validate backoff strategy
	if r.BackoffStrategy != nil {
		if err := r.BackoffStrategy.Validate(); err != nil {
			return fmt.Errorf("invalid backoff strategy: %w", err)
		}
	}

	return nil
}

// Validate validates the backoff strategy
func (bs *BackoffStrategy) Validate() error {
	validTypes := map[string]bool{
		"exponential": true,
		"fixed":       true,
	}

	if !validTypes[bs.Type] {
		return fmt.Errorf("invalid backoff type: %s", bs.Type)
	}

	if bs.BaseInterval == "" {
		return fmt.Errorf("base_interval is required")
	}

	if _, err := time.ParseDuration(bs.BaseInterval); err != nil {
		return fmt.Errorf("invalid base_interval: %w", err)
	}

	if bs.MaxInterval != "" {
		if _, err := time.ParseDuration(bs.MaxInterval); err != nil {
			return fmt.Errorf("invalid max_interval: %w", err)
		}
	}

	if bs.Type == "exponential" && bs.Multiplier <= 0 {
		return fmt.Errorf("multiplier must be positive for exponential backoff")
	}

	return nil
}

// Validate validates the circuit breaker settings
func (cb *CircuitBreaker) Validate() error {
	if cb.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be positive")
	}

	if cb.MaxPendingRequests <= 0 {
		return fmt.Errorf("max_pending_requests must be positive")
	}

	if cb.MaxRequests <= 0 {
		return fmt.Errorf("max_requests must be positive")
	}

	if cb.ErrorThresholdPercent < 0 || cb.ErrorThresholdPercent > 100 {
		return fmt.Errorf("error_threshold_percent must be between 0 and 100")
	}

	if cb.SleepWindow == "" {
		return fmt.Errorf("sleep_window is required")
	}

	if _, err := time.ParseDuration(cb.SleepWindow); err != nil {
		return fmt.Errorf("invalid sleep_window: %w", err)
	}

	return nil
}

// IsOpen checks if the circuit breaker should be open based on metrics
func (cb *CircuitBreaker) IsOpen(metrics *ServiceMetrics) bool {
	if metrics == nil {
		return false
	}

	// Check if error rate exceeds threshold
	if metrics.RequestCount > 0 {
		errorRate := (float64(metrics.ErrorCount) / float64(metrics.RequestCount)) * 100
		if errorRate >= cb.ErrorThresholdPercent {
			return true
		}
	}

	return metrics.CircuitBreakerOpen
}

// Validate validates the timeout policy
func (t *TimeoutPolicy) Validate() error {
	if t.ConnectionTimeout == "" {
		return fmt.Errorf("connection_timeout is required")
	}

	if _, err := time.ParseDuration(t.ConnectionTimeout); err != nil {
		return fmt.Errorf("invalid connection_timeout: %w", err)
	}

	if t.RequestTimeout == "" {
		return fmt.Errorf("request_timeout is required")
	}

	if _, err := time.ParseDuration(t.RequestTimeout); err != nil {
		return fmt.Errorf("invalid request_timeout: %w", err)
	}

	if t.IdleTimeout != "" {
		if _, err := time.ParseDuration(t.IdleTimeout); err != nil {
			return fmt.Errorf("invalid idle_timeout: %w", err)
		}
	}

	return nil
}

// Validate validates the rate limit policy
func (rl *RateLimitPolicy) Validate() error {
	if rl.RequestsPerUnit <= 0 {
		return fmt.Errorf("requests_per_unit must be positive")
	}

	validUnits := map[string]bool{
		"second": true,
		"minute": true,
		"hour":   true,
	}

	if !validUnits[rl.Unit] {
		return fmt.Errorf("invalid unit: %s", rl.Unit)
	}

	if rl.BurstSize < 0 {
		return fmt.Errorf("burst_size cannot be negative")
	}

	if rl.FillRate < 0 {
		return fmt.Errorf("fill_rate cannot be negative")
	}

	return nil
}

// Validate validates the health check policy
func (hc *HealthCheckPolicy) Validate() error {
	validProtocols := map[string]bool{
		"HTTP": true,
		"TCP":  true,
		"gRPC": true,
	}

	if !validProtocols[hc.Protocol] {
		return fmt.Errorf("invalid protocol: %s", hc.Protocol)
	}

	if hc.Protocol == "HTTP" && hc.Path == "" {
		return fmt.Errorf("path is required for HTTP health checks")
	}

	if hc.Interval == "" {
		return fmt.Errorf("interval is required")
	}

	if _, err := time.ParseDuration(hc.Interval); err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	if hc.Timeout == "" {
		return fmt.Errorf("timeout is required")
	}

	if _, err := time.ParseDuration(hc.Timeout); err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	if hc.HealthyThreshold <= 0 {
		return fmt.Errorf("healthy_threshold must be positive")
	}

	if hc.UnhealthyThreshold <= 0 {
		return fmt.Errorf("unhealthy_threshold must be positive")
	}

	return nil
}

// Validate validates the fault injection policy
func (fi *FaultInjectionPolicy) Validate() error {
	if fi.Delay != nil {
		if err := fi.Delay.Validate(); err != nil {
			return fmt.Errorf("invalid delay config: %w", err)
		}
	}

	if fi.Abort != nil {
		if err := fi.Abort.Validate(); err != nil {
			return fmt.Errorf("invalid abort config: %w", err)
		}
	}

	return nil
}

// Validate validates the fault delay configuration
func (fd *FaultDelay) Validate() error {
	if fd.Percentage < 0 || fd.Percentage > 100 {
		return fmt.Errorf("percentage must be between 0 and 100")
	}

	if fd.FixedDelay == "" {
		return fmt.Errorf("fixed_delay is required")
	}

	if _, err := time.ParseDuration(fd.FixedDelay); err != nil {
		return fmt.Errorf("invalid fixed_delay: %w", err)
	}

	return nil
}

// Validate validates the fault abort configuration
func (fa *FaultAbort) Validate() error {
	if fa.Percentage < 0 || fa.Percentage > 100 {
		return fmt.Errorf("percentage must be between 0 and 100")
	}

	if fa.HTTPStatus < 100 || fa.HTTPStatus >= 600 {
		return fmt.Errorf("http_status must be a valid HTTP status code")
	}

	return nil
}

// DefaultTrafficPolicy returns a default traffic policy
func DefaultTrafficPolicy() *TrafficPolicy {
	return &TrafficPolicy{
		Name: "default",
		LoadBalancing: &LoadBalancingPolicy{
			Algorithm: "ROUND_ROBIN",
		},
		Retry: &RetryPolicy{
			Attempts:      3,
			PerTryTimeout: "5s",
			RetryOn:       []string{"5xx", "reset", "connect-failure"},
			BackoffStrategy: &BackoffStrategy{
				Type:         "exponential",
				BaseInterval: "100ms",
				MaxInterval:  "10s",
				Multiplier:   2.0,
			},
		},
		CircuitBreaker: &CircuitBreaker{
			MaxConnections:        1024,
			MaxPendingRequests:    1024,
			MaxRequests:           1024,
			MaxRetries:            3,
			ErrorThresholdPercent: 50.0,
			SleepWindow:           "30s",
			ConsecutiveErrors:     5,
		},
		Timeout: &TimeoutPolicy{
			ConnectionTimeout: "10s",
			RequestTimeout:    "30s",
			IdleTimeout:       "60s",
		},
		HealthCheck: &HealthCheckPolicy{
			Protocol:           "HTTP",
			Path:               "/health",
			Interval:           "10s",
			Timeout:            "5s",
			HealthyThreshold:   2,
			UnhealthyThreshold: 3,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// HighAvailabilityPolicy returns a high availability traffic policy
func HighAvailabilityPolicy() *TrafficPolicy {
	return &TrafficPolicy{
		Name: "high-availability",
		LoadBalancing: &LoadBalancingPolicy{
			Algorithm: "LEAST_REQUEST",
		},
		Retry: &RetryPolicy{
			Attempts:      5,
			PerTryTimeout: "3s",
			RetryOn:       []string{"5xx", "reset", "connect-failure", "retriable-4xx"},
			BackoffStrategy: &BackoffStrategy{
				Type:         "exponential",
				BaseInterval: "50ms",
				MaxInterval:  "5s",
				Multiplier:   2.0,
			},
		},
		CircuitBreaker: &CircuitBreaker{
			MaxConnections:        2048,
			MaxPendingRequests:    2048,
			MaxRequests:           2048,
			MaxRetries:            5,
			ErrorThresholdPercent: 30.0,
			SleepWindow:           "15s",
			ConsecutiveErrors:     3,
		},
		Timeout: &TimeoutPolicy{
			ConnectionTimeout: "5s",
			RequestTimeout:    "15s",
			IdleTimeout:       "120s",
		},
		HealthCheck: &HealthCheckPolicy{
			Protocol:           "HTTP",
			Path:               "/health",
			Interval:           "5s",
			Timeout:            "3s",
			HealthyThreshold:   2,
			UnhealthyThreshold: 2,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// PerformancePolicy returns a performance-optimized traffic policy
func PerformancePolicy() *TrafficPolicy {
	return &TrafficPolicy{
		Name: "performance",
		LoadBalancing: &LoadBalancingPolicy{
			Algorithm: "RANDOM",
		},
		Retry: &RetryPolicy{
			Attempts:      2,
			PerTryTimeout: "2s",
			RetryOn:       []string{"5xx", "reset"},
			BackoffStrategy: &BackoffStrategy{
				Type:         "fixed",
				BaseInterval: "100ms",
			},
		},
		CircuitBreaker: &CircuitBreaker{
			MaxConnections:        4096,
			MaxPendingRequests:    4096,
			MaxRequests:           4096,
			MaxRetries:            2,
			ErrorThresholdPercent: 60.0,
			SleepWindow:           "45s",
			ConsecutiveErrors:     10,
		},
		Timeout: &TimeoutPolicy{
			ConnectionTimeout: "3s",
			RequestTimeout:    "10s",
			IdleTimeout:       "30s",
		},
		HealthCheck: &HealthCheckPolicy{
			Protocol:           "HTTP",
			Path:               "/health",
			Interval:           "15s",
			Timeout:            "5s",
			HealthyThreshold:   3,
			UnhealthyThreshold: 5,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
