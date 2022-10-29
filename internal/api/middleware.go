package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
)

var (
	AccessCookieName   = "CF_Authorization"
	AccessIdentityPath = "/cdn-cgi/access/get-identity"
	AccessCertsPath    = "/cdn-cgi/access/certs"
)

type AccessIdentity struct {
	Email              string `json:"email"`
	UserUUID           string `json:"user_uuid"`
	AccountID          string `json:"account_id"`
	ServiceTokenID     string `json:"service_token_id"`
	ServiceTokenStatus bool   `json:"service_token_status"`
}

type AccessClient struct {
	verifier   *oidc.IDTokenVerifier
	httpClient *http.Client
	domain     string
	policyAud  string
}

func (a AccessClient) Verify(ctx context.Context, jwt string) (*oidc.IDToken, error) {
	return a.verifier.Verify(ctx, jwt)
}

func (a AccessClient) GetIdentity(ctx context.Context, cfAuthorization *http.Cookie) (*AccessIdentity, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s%s", a.domain, AccessIdentityPath), nil)
	if err != nil {
		log.Error().Err(err).Msg("creating access identity request")
		return nil, err
	}
	req.AddCookie(cfAuthorization)
	res, err := a.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("sending access identity request")
		return nil, err
	}
	if res.StatusCode >= http.StatusBadRequest {
		e := struct {
			Error string `json:"err"`
		}{}
		err = json.NewDecoder(res.Body).Decode(&e)
		if err == nil {
			log.Error().Str("status", res.Status).Int("code", res.StatusCode).Str("err", e.Error).Msg("received access identity request")
		}
		return nil, fmt.Errorf("get access identity: %d %s", res.StatusCode, res.Status)
	}
	identity := &AccessIdentity{}
	err = json.NewDecoder(res.Body).Decode(identity)
	if err != nil {
		log.Error().Err(err).Msg("decoding access identity")
		return nil, err
	}
	return identity, nil
}

func NewAccessClient(teamDomain, policyAUD string) AccessClient {
	certsURL := fmt.Sprintf("https://%s%s", teamDomain, AccessCertsPath)

	config := &oidc.Config{
		ClientID: policyAUD,
	}
	keySet := oidc.NewRemoteKeySet(context.Background(), certsURL)
	verifier := oidc.NewVerifier(fmt.Sprintf("https://%s", teamDomain), keySet, config)

	if policyAUD == "" && teamDomain == "" {
		log.Debug().Msg("no cloudflare access policy aud set, skipping authentication")
	} else {
		log.Debug().Str("team_domain", teamDomain).Str("policy_aud", policyAUD).Msg("cloudflare access configured")

	}
	return AccessClient{
		verifier:   verifier,
		httpClient: &http.Client{Timeout: time.Second * 3},
		domain:     teamDomain,
		policyAud:  policyAUD,
	}
}

func CloudflareAccessVerifier(client AccessClient) func(next http.Handler) http.Handler {
	if client.policyAud == "" && client.domain == "" {
		return func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(fn)
		}
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			// Make sure that the incoming request has our token header
			//  Could also look in the cookies for CF_AUTHORIZATION
			accessJWT := r.Header.Get("Cf-Access-Jwt-Assertion")
			if accessJWT == "" {
				log.Debug().Str("jwt", accessJWT).Msg("couldn't get authorization token")
				writeErr(w, nil, ErrUnauthorized)
				return
			}

			// Verify the access token
			_, err := client.Verify(r.Context(), accessJWT)
			if err != nil {
				log.Debug().Err(err).Msg("invalid token")
				writeErr(w, nil, ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
