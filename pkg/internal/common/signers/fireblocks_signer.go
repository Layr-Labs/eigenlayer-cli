package signers

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/fireblocks"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"time"
)

type CreateTransactionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type TransactionSignedMessageSignature struct {
	R       string `json:"r"`
	S       string `json:"s"`
	V       int    `json:"v"`
	FullSig string `json:"fullSig"`
}
type TransactionSignedMessage struct {
	PublicKey string                            `json:"publicKey"`
	Signature TransactionSignedMessageSignature `json:"signature"`
}
type TransactionResponse struct {
	Status         string                     `json:"status"`
	SignedMessages []TransactionSignedMessage `json:"signedMessages"`
}

type FireblocksSigner struct {
	client           *FireblocksClient
	chainID          uint64
	vaultAccountName string
	account          *fireblocks.VaultAccount
}

func NewFireblocksSigner(client *FireblocksClient, chainID uint64, vaultAccountName string) *FireblocksSigner {
	return &FireblocksSigner{client: client, chainID: chainID, vaultAccountName: vaultAccountName}
}

func (m *FireblocksSigner) getTxStatus(ctx context.Context, txID string) (*TransactionResponse, error) {
	// TODO: make this ticker adjustable
	queryTicker := time.NewTicker(2 * time.Second)
	defer queryTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, errors.Join(errors.New("context done before tx was mined"), ctx.Err())
		case <-queryTicker.C:
			receipt, err := m.queryTxStatus(txID)
			if err != nil {
				return nil, err
			}
			if receipt != nil {
				return receipt, nil
			}
		}
	}
}

func (m *FireblocksSigner) queryTxStatus(txID string) (*TransactionResponse, error) {
	var resp TransactionResponse
	err := m.makeRequest(context.Background(), http.MethodGet, fmt.Sprintf("/v1/transactions/%s", txID), nil, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Status == "COMPLETED" {
		return &resp, nil
	}
	switch resp.Status {
	case "FAILED":
		return nil, fmt.Errorf("transaction failed with status: %s", resp.Status)
	case "BLOCKED":
		return nil, fmt.Errorf("transaction failed with status: %s", resp.Status)
	case "CANCELLED":
		return nil, fmt.Errorf("transaction failed with status: %s", resp.Status)
	case "REJECTED":
		return nil, fmt.Errorf("transaction failed with status: %s", resp.Status)
	}
	m.client.logger.Info(
		fmt.Sprintf(
			"transaction not yet broadcasted: the Fireblocks transaction %s is in status %s",
			txID,
			resp.Status,
		),
	)
	return nil, nil
}

func (m *FireblocksSigner) getAccount(ctx context.Context) (*fireblocks.VaultAccount, error) {
	if m.account == nil {
		accounts, err := m.ListVaultAccounts(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing vault accounts: %w", err)
		}
		for i, a := range accounts {
			if a.Name == m.vaultAccountName {
				m.account = &accounts[i]
				break
			}
		}
	}
	return m.account, nil
}

func (m *FireblocksSigner) ListVaultAccounts(ctx context.Context) ([]fireblocks.VaultAccount, error) {
	var accounts []fireblocks.VaultAccount
	type paging struct {
		Before string `json:"before"`
		After  string `json:"after"`
	}
	var response struct {
		Accounts []fireblocks.VaultAccount `json:"accounts"`
		Paging   paging                    `json:"paging"`
	}
	p := paging{}
	next := true
	for next {
		u, err := url.Parse("/v1/vault/accounts_paged")
		if err != nil {
			return accounts, fmt.Errorf("error parsing URL: %w", err)
		}
		q := u.Query()
		q.Set("before", p.Before)
		q.Set("after", p.After)
		u.RawQuery = q.Encode()
		err = m.makeRequest(ctx, "GET", u.String(), nil, &response)
		if err != nil {
			return accounts, fmt.Errorf("error making request: %w", err)
		}

		accounts = append(accounts, response.Accounts...)
		p = response.Paging
		if p.After == "" {
			next = false
		}
	}

	return accounts, nil
}

func (m *FireblocksSigner) Sign(data []byte) ([]byte, error) {
	digestHashHex := hex.EncodeToString(data)
	m.client.logger.Debug(fmt.Sprintf("Signing digest hash using fireblocks: %s", digestHashHex))

	assetID, ok := fireblocks.AssetIDByChain[m.chainID]
	if !ok {
		return nil, fmt.Errorf("unsupported chain %d", m.chainID)
	}
	account, err := m.getAccount(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting account: %w", err)
	}
	foundAsset := false
	for _, a := range account.Assets {
		if a.ID == assetID {
			if a.Available == "0" {
				return nil, errors.New("insufficient funds")
			}
			foundAsset = true
			break
		}
	}
	if !foundAsset {
		return nil, fmt.Errorf("asset %s not found in account %s", assetID, m.vaultAccountName)
	}

	payload := map[string]any{
		"assetId":   assetID,
		"operation": "RAW",
		"source": map[string]any{
			"type": "VAULT_ACCOUNT",
			"id":   account.ID,
		},
		"note": "",
		"extraParameters": map[string]any{
			"rawMessageData": map[string]any{
				"messages": []map[string]any{
					{
						"content":           digestHashHex,
						"bip44addressIndex": 0,
					},
				},
			},
		},
	}
	var response CreateTransactionResponse
	err = m.makeRequest(context.Background(), "POST", "/v1/transactions", payload, &response)
	if err != nil {
		return nil, err
	}

	status, err2 := m.getTxStatus(context.Background(), response.ID)
	if err2 != nil {
		return nil, err2
	}

	sig := status.SignedMessages[0].Signature

	encodedSig := sig.FullSig + fmt.Sprintf("0%x", int64(sig.V))
	decodedSig, err := hex.DecodeString(encodedSig)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %v", err)
	}

	return decodedSig, nil
}

func (m *FireblocksSigner) makeRequest(
	ctx context.Context,
	httpMethod string,
	path string,
	payload any,
	response any,
) error {
	resp, err := m.client.makeRequest(ctx, httpMethod, path, payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp, response)
}

type FireblocksClient struct {
	apiKey     string
	privateKey *rsa.PrivateKey
	baseURL    string
	timeout    time.Duration
	client     *http.Client
	logger     logging.Logger
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewFireblocksClient(
	apiKey string,
	secretKey []byte,
	baseURL string,
	timeout time.Duration,
	logger logging.Logger,
) (*FireblocksClient, error) {
	c := http.Client{Timeout: timeout}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(secretKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing RSA private key: %w", err)
	}

	return &FireblocksClient{
		apiKey:     apiKey,
		privateKey: privateKey,
		baseURL:    baseURL,
		timeout:    timeout,
		client:     &c,
		logger:     logger,
	}, nil
}

// signJwt signs a JWT token for the Fireblocks API
// mostly copied from the Fireblocks example:
// https://github.com/fireblocks/developers-hub/blob/main/authentication_examples/go/test.go
func (f *FireblocksClient) signJwt(path string, bodyJson interface{}, durationSeconds int64) (string, error) {
	nonce := uuid.New().String()
	now := time.Now().Unix()
	expiration := now + durationSeconds

	bodyBytes, err := json.Marshal(bodyJson)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	h := sha256.New()
	h.Write(bodyBytes)
	hashed := h.Sum(nil)

	claims := jwt.MapClaims{
		"uri":      path,
		"nonce":    nonce,
		"iat":      now,
		"exp":      expiration,
		"sub":      f.apiKey,
		"bodyHash": hex.EncodeToString(hashed),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(f.privateKey)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, nil
}

// makeRequest makes a request to the Fireblocks API
// mostly copied from the Fireblocks example:
// https://github.com/fireblocks/developers-hub/blob/main/authentication_examples/go/test.go
func (f *FireblocksClient) makeRequest(_ context.Context, method, path string, body interface{}) ([]byte, error) {
	// remove query parameters from path and join with baseURL
	pathURI, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}
	query := pathURI.Query()
	pathURI.RawQuery = ""
	urlStr, err := url.JoinPath(f.baseURL, pathURI.String())
	if err != nil {
		return nil, fmt.Errorf("error joining URL path with %s and %s: %w", f.baseURL, path, err)
	}
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}
	// add query parameters back to path
	parsedUrl.RawQuery = query.Encode()
	f.logger.Debug("making request to Fireblocks", "method", method, "url", parsedUrl.String())
	var reqBodyBytes []byte
	if body != nil {
		var err error
		reqBodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
	}

	token, err := f.signJwt(path, body, int64(f.timeout.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("error signing JWT: %w", err)
	}

	req, err := http.NewRequest(method, parsedUrl.String(), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-API-KEY", f.apiKey)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			f.logger.Error("error closing HTTP response body", "error", err)
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		err = json.Unmarshal(respBody, &errResp)
		if err != nil {
			return nil, fmt.Errorf("error parsing error response: %w", err)
		}
		return nil, fmt.Errorf(
			"error response (%d) from Fireblocks with code %d: %s",
			resp.StatusCode,
			errResp.Code,
			errResp.Message,
		)
	}

	return respBody, nil
}
