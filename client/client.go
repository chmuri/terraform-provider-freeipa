package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/jcmturner/gokrb5/v8/spnego"
)

// RPCRequest standard JSON-RPC request for FreeIPA
type RPCRequest struct {
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Version string        `json:"jsonrpc"`
}

// RPCResponse standard JSON-RPC response from FreeIPA
type RPCResponse struct {
	ID        int             `json:"id"`
	Version   string          `json:"jsonrpc"`
	Error     *RPCError       `json:"error"`
	Principal string          `json:"principal"`
	Result    json.RawMessage `json:"result"`
}

// RPCError represents a JSON-RPC error from FreeIPA
type RPCError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Name    string                 `json:"name"`
	Data    map[string]interface{} `json:"data"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("FreeIPA API error %d (%s): %s", e.Code, e.Name, e.Message)
}

// AuthMethod represents chosen authentication method
type AuthMethod int

const (
	AuthPassword AuthMethod = iota
	AuthKerberosKeytab
)

// Config holds client configurations
type Config struct {
	Host       string
	Insecure   bool
	AuthMethod AuthMethod

	// For AuthPassword
	Username string
	Password string

	// For AuthKerberosKeytab
	KeytabPath string
	Realm      string // e.g., "DEMO.FREEIPA.ORG"
	Krb5Conf   string // e.g., "/etc/krb5.conf"
}

// Client is FreeIPA JSON-RPC client
type Client struct {
	cfg        *Config
	httpClient *http.Client
	mu         sync.Mutex
}

// NewClient creates a new FreeIPA client
func NewClient(cfg *Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Insecure,
		},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if strings.HasPrefix(addr, "ipa.test.local:") {
				addr = "10.5.0.10:" + strings.Split(addr, ":")[1]
			}
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}

	httpClient := &http.Client{
		Transport: tr,
		Jar:       jar,
	}

	return &Client{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

// Login performs authentication based on the chosen method
func (c *Client) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.cfg.AuthMethod {
	case AuthPassword:
		return c.loginPassword()
	case AuthKerberosKeytab:
		return c.loginKerberos()
	default:
		return fmt.Errorf("unsupported auth method")
	}
}

// loginPassword performs password authentication using login_password endpoint
func (c *Client) loginPassword() error {
	loginURL := fmt.Sprintf("https://%s/ipa/session/login_password", c.cfg.Host)

	form := url.Values{}
	form.Set("user", c.cfg.Username)
	form.Set("password", c.cfg.Password)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", fmt.Sprintf("https://%s/ipa", c.cfg.Host))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("password login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		rejection := resp.Header.Get("X-IPA-Rejection-Reason")
		if rejection != "" {
			return fmt.Errorf("auth rejected: %s", rejection)
		}
		return fmt.Errorf("unauthorized")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("password login failed with HTTP status: %d", resp.StatusCode)
	}

	return nil
}

// loginKerberos performs Kerberos SPNEGO authentication using login_kerberos endpoint
func (c *Client) loginKerberos() error {
	krb5ConfPath := c.cfg.Krb5Conf
	if krb5ConfPath == "" {
		krb5ConfPath = "/etc/krb5.conf"
	}

	krbCfg, err := config.Load(krb5ConfPath)
	if err != nil {
		return fmt.Errorf("failed to load krb5.conf: %w", err)
	}

	kt, err := keytab.Load(c.cfg.KeytabPath)
	if err != nil {
		return fmt.Errorf("failed to load keytab: %w", err)
	}

	krbClient := client.NewWithKeytab(c.cfg.Username, c.cfg.Realm, kt, krbCfg, client.DisablePAFXFAST(true))
	if err := krbClient.Login(); err != nil {
		return fmt.Errorf("kerberos kinit failed: %w", err)
	}

	loginURL := fmt.Sprintf("https://%s/ipa/session/login_kerberos", c.cfg.Host)
	req, err := http.NewRequest("POST", loginURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Referer", fmt.Sprintf("https://%s/ipa", c.cfg.Host))

	spn := fmt.Sprintf("HTTP/%s", c.cfg.Host)
	if err := spnego.SetSPNEGOHeader(krbClient, req, spn); err != nil {
		return fmt.Errorf("failed to generate SPNEGO token: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("kerberos login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kerberos login failed with HTTP status: %d", resp.StatusCode)
	}

	return nil
}

// Call executes a FreeIPA JSON-RPC command
func (c *Client) Call(ctx context.Context, method string, args []string, options map[string]interface{}, out interface{}) error {
	if options == nil {
		options = make(map[string]interface{})
	}
	if _, ok := options["version"]; !ok {
		options["version"] = "2.231"
	}

	params := []interface{}{args, options}
	reqBody := RPCRequest{
		ID:      1,
		Method:  method,
		Params:  params,
		Version: "2.0",
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	executeCall := func() (*http.Response, error) {
		apiURL := fmt.Sprintf("https://%s/ipa/session/json", c.cfg.Host)
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Referer", fmt.Sprintf("https://%s/ipa", c.cfg.Host))

		return c.httpClient.Do(req)
	}

	resp, err := executeCall()
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		if err := c.Login(); err != nil {
			return fmt.Errorf("session expired, re-login failed: %w", err)
		}

		resp, err = executeCall()
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	var rpcResp RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return err
	}

	if rpcResp.Error != nil {
		return rpcResp.Error
	}

	if out != nil {
		if err := json.Unmarshal(rpcResp.Result, out); err != nil {
			return fmt.Errorf("failed to unmarshal response result: %w", err)
		}
	}

	return nil
}
