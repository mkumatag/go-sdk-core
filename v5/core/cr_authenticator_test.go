// +build all slow auth

package core

// (C) Copyright IBM Corp. 2019.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

const (
	// To enable debug logging during test execution, set this to "LevelDebug"
	craTestLogLevel       LogLevel = LevelError
	craMockCRTokenFile    string   = "../resources/cr-token.txt"
	craMockIAMProfileName string   = "iam-user-123"
	craMockIAMProfileID   string   = "iam-id-123"
	craMockClientID       string   = "client-id-1"
	craMockClientSecret   string   = "client-secret-1"
	craTestCRToken1       string   = "cr-token-1"
	craTestAccessToken1   string   = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	craTestAccessToken2   string   = "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	craTestRefreshToken   string   = "Xj7Gle500MachEOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
)

// Struct that models the request body for the "create_access_token" operation
type instanceIdentityTokenPrototype struct {
	ExpiresIn int `json:"expires_in"`
}

func TestCraCtorErrors(t *testing.T) {
	var err error
	var auth *ComputeResourceAuthenticator

	// Error: missing IAMProfileName and IBMProfileID.
	auth, err = NewComputeResourceAuthenticator("", "", "", "", "", "", "", false, "", nil)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientID.
	auth, err = NewComputeResourceAuthenticator("", "", craMockIAMProfileName, "", "", "", "client-secret", false, "", nil)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientSecret.
	auth, err = NewComputeResourceAuthenticator("", "", "", "iam-id-123", "", "client-id", "", false, "", nil)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}

func TestCraCtorSuccess(t *testing.T) {
	var err error
	var auth *ComputeResourceAuthenticator
	var expectedHeaders = map[string]string{
		"header1": "value1",
	}

	// Success - only required params
	// 1. only IAMProfileName
	auth, err = NewComputeResourceAuthenticator("", "", craMockIAMProfileName, "", "", "", "", false, "", nil)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRAUTH, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.InstanceMetadataServiceURL)
	assert.Equal(t, craMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// 2. only IAMProfileID
	auth, err = NewComputeResourceAuthenticator("", "", "", craMockIAMProfileID, "", "", "", false, "", nil)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRAUTH, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.InstanceMetadataServiceURL)
	assert.Equal(t, "", auth.IAMProfileName)
	assert.Equal(t, craMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// Success - all parameters
	auth, err = NewComputeResourceAuthenticator("cr-token-file", "http://1.1.1.1", craMockIAMProfileName, craMockIAMProfileID,
		defaultIamTokenServerEndpoint, "client-id", "client-secret", true, "scope1", expectedHeaders)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRAUTH, auth.AuthenticationType())
	assert.Equal(t, "cr-token-file", auth.CRTokenFilename)
	assert.Equal(t, "http://1.1.1.1", auth.InstanceMetadataServiceURL)
	assert.Equal(t, craMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, craMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, defaultIamTokenServerEndpoint, auth.URL)
	assert.Equal(t, "client-id", auth.ClientID)
	assert.Equal(t, "client-secret", auth.ClientSecret)
	assert.Equal(t, true, auth.DisableSSLVerification)
	assert.Equal(t, expectedHeaders, auth.Headers)
}

func TestCraCtorFromMapErrors(t *testing.T) {
	var err error
	var auth *ComputeResourceAuthenticator
	var configProps map[string]string

	// Error: nil config map
	auth, err = newComputeResourceAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing IAMProfileName and IAMProfileID
	configProps = map[string]string{}
	auth, err = newComputeResourceAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientID.
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_NAME: craMockIAMProfileName,
		PROPNAME_CLIENT_SECRET:    "client-secret",
	}
	auth, err = newComputeResourceAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientSecret.
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_ID: "iam-id-123",
		PROPNAME_CLIENT_ID:      "client-id",
	}
	auth, err = newComputeResourceAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}
func TestCraCtorFromMapSuccess(t *testing.T) {
	var err error
	var auth *ComputeResourceAuthenticator
	var configProps map[string]string

	// Success - only required params
	// 1. only IAMProfileName
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_NAME: craMockIAMProfileName,
	}
	auth, err = newComputeResourceAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRAUTH, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.InstanceMetadataServiceURL)
	assert.Equal(t, craMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// 2. only IAMProfileID
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_ID: craMockIAMProfileID,
	}
	auth, err = newComputeResourceAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRAUTH, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.InstanceMetadataServiceURL)
	assert.Equal(t, "", auth.IAMProfileName)
	assert.Equal(t, craMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// Success - all params
	configProps = map[string]string{
		PROPNAME_CRTOKEN_FILENAME:              "cr-token-file",
		PROPNAME_INSTANCE_METADATA_SERVICE_URL: "http://1.1.1.1",
		PROPNAME_IAM_PROFILE_NAME:              craMockIAMProfileName,
		PROPNAME_IAM_PROFILE_ID:                "iam-id-123",
		PROPNAME_AUTH_URL:                      defaultIamTokenServerEndpoint,
		PROPNAME_CLIENT_ID:                     "client-id",
		PROPNAME_CLIENT_SECRET:                 "client-secret",
		PROPNAME_AUTH_DISABLE_SSL:              "true",
		PROPNAME_SCOPE:                         "scope1",
	}
	auth, err = newComputeResourceAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRAUTH, auth.AuthenticationType())
	assert.Equal(t, "cr-token-file", auth.CRTokenFilename)
	assert.Equal(t, "http://1.1.1.1", auth.InstanceMetadataServiceURL)
	assert.Equal(t, craMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, "iam-id-123", auth.IAMProfileID)
	assert.Equal(t, defaultIamTokenServerEndpoint, auth.URL)
	assert.Equal(t, "client-id", auth.ClientID)
	assert.Equal(t, "client-secret", auth.ClientSecret)
	assert.Equal(t, true, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)
}

// startMockServer will start a mock server endpoint that supports both the
// Instance Metadata Service and IAM operations that we'll need to call.
func startMockServer(t *testing.T) *httptest.Server {
	// Create the mock server.
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		operationPath := req.URL.EscapedPath()
		method := req.Method

		if operationPath == "/instance_identity/v1/token" {
			// If this is an invocation of the IMDS "create_access_token" operation,
			// then validate it a bit and then send back a good response.
			assert.Equal(t, "PUT", method)
			assert.Equal(t, imdsVersionDate, req.URL.Query()["version"][0])
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Accept"))
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Content-Type"))
			assert.Equal(t, imdsMetadataFlavor, req.Header.Get("Metadata-Flavor"))

			// Read and unmarshal the request body.
			requestBody := &instanceIdentityTokenPrototype{}
			_ = json.NewDecoder(req.Body).Decode(requestBody)
			defer req.Body.Close()

			assert.NotNil(t, requestBody)
			assert.Equal(t, crtokenLifetime, requestBody.ExpiresIn)

			res.WriteHeader(http.StatusOK)
			fmt.Fprintf(res, `{"access_token":"%s"}`, craTestCRToken1)
		} else if operationPath == "/identity/token" {
			// If this is an invocation of the IAM "get_token" operation,
			// then validate it a bit and then send back a good response.
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Accept"))
			assert.Equal(t, FORM_URL_ENCODED_HEADER, req.Header.Get("Content-Type"))
			assert.Equal(t, craTestCRToken1, req.FormValue("cr_token"))
			assert.Equal(t, iamGrantTypeCRToken, req.FormValue("grant_type"))

			iamProfileID := req.FormValue("profile_id")
			iamProfileName := req.FormValue("profile_name")
			assert.True(t, iamProfileName != "" || iamProfileID != "")

			// Assume that we'll return a 200 OK status code.
			statusCode := http.StatusOK

			// This is the access token we'll send back in the mock response.
			// We'll default to token 1, then see if the caller asked for token 2
			// via the scope setting below.
			accessToken := craTestAccessToken1

			// We'll use the scope value to control the behavior of this mock endpoint so that we can force
			// certain things to happen:
			// 1. whether to return the first or second access token.
			// 2. whether we should validate the basic-auth header.
			// 3. whether we should return a bad status code.
			// Yes, this is kinda subversive but sometimes we need to be creative on these big jobs :)
			scope := req.FormValue("scope")

			if scope == "send-second-token" {
				accessToken = craTestAccessToken2
			} else if scope == "check-basic-auth" {
				username, password, ok := req.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, craMockClientID, username)
				assert.Equal(t, craMockClientSecret, password)
			} else if scope == "check-user-headers" {
				assert.Equal(t, "Value-1", req.Header.Get("User-Header-1"))
				assert.Equal(t, "iam.cloud.ibm.com", req.Host)
			} else if scope == "status-bad-request" {
				statusCode = http.StatusBadRequest
			} else if scope == "status-unauthorized" {
				statusCode = http.StatusUnauthorized
			} else if scope == "sleep" {
				time.Sleep(3 * time.Second)
			}

			expiration := GetCurrentTime() + 3600
			res.WriteHeader(statusCode)
			switch statusCode {
			case http.StatusOK:
				fmt.Fprintf(res, `{"access_token": "%s", "token_type": "Bearer", "expires_in": 3600, "expiration": %d, "refresh_token": "%s"}`,
					accessToken, expiration, craTestRefreshToken)
			case http.StatusBadRequest:
				fmt.Fprintf(res, `Sorry, bad request!`)

			case http.StatusUnauthorized:
				fmt.Fprintf(res, `Sorry, you are not authorized!`)
			}
		} else {
			assert.Fail(t, "unknown operation path: "+operationPath)
		}
	}))
	return server
}

func TestCraRetrieveCRTokenFromFileSuccess(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	// Set the authenticator to read the CR token from our mock file.
	auth := &ComputeResourceAuthenticator{
		CRTokenFilename: craMockCRTokenFile,
	}
	crToken, err := auth.readCRTokenFromFile()
	assert.Nil(t, err)
	assert.Equal(t, craTestCRToken1, crToken)
}

func TestCraRetrieveCRTokenFromFileFail(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	// Use a non-existent cr token file.
	auth := &ComputeResourceAuthenticator{
		CRTokenFilename: "bogus-cr-token-file",
	}
	crToken, err := auth.readCRTokenFromFile()
	assert.NotNil(t, err)
	assert.Equal(t, "", crToken)
	t.Logf("Expected error: %s", err.Error())
}

func TestCraRetrieveCRTokenFromIMDSSuccess(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	// Set up the authenticator to use the mock server.
	auth := &ComputeResourceAuthenticator{
		InstanceMetadataServiceURL: server.URL,
	}
	crToken, err := auth.retrieveCRTokenFromIMDS()
	assert.Nil(t, err)
	assert.Equal(t, craTestCRToken1, crToken)
}

func TestCraRetrieveCRTokenFromIMDSFail(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	// Set up the authenticator to use a bogus IMDS endpoint.
	auth := &ComputeResourceAuthenticator{
		InstanceMetadataServiceURL: "http://bogus.imds.endpoint",
	}
	crToken, err := auth.retrieveCRTokenFromIMDS()
	assert.NotNil(t, err)
	assert.Equal(t, "", crToken)
	t.Logf("Expected error: %s", err.Error())
}

func TestCraGetTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	auth := &ComputeResourceAuthenticator{
		CRTokenFilename: craMockCRTokenFile,
		IAMProfileName:  craMockIAMProfileName,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Verify that we initially have no token data cached on the authenticator.
	assert.Nil(t, auth.getTokenData())

	// Force the first fetch and verify we got the first access token.
	var accessToken string
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)

	// Verify that the access token was returned by GetToken() and also
	// stored in the authenticator's tokenData field as well.
	assert.NotNil(t, auth.getTokenData())
	assert.Equal(t, craTestAccessToken1, accessToken)
	assert.Equal(t, craTestAccessToken1, auth.getTokenData().AccessToken)

	// Call GetToken() again and verify that we get the cached value.
	// Note: we'll Set Scope so that if the IAM operation is actually called again,
	// we'll receive the second access token.  We don't want the IAM operation called again yet.
	auth.Scope = "send-second-token"
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)

	// Force expiration and verify that GetToken() fetched the second access token.
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	auth.IAMProfileName = ""
	auth.IAMProfileID = craMockIAMProfileID
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, auth.getTokenData())
	assert.Equal(t, craTestAccessToken2, accessToken)
	assert.Equal(t, craTestAccessToken2, auth.getTokenData().AccessToken)
}

func TestCraRequestTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	auth := &ComputeResourceAuthenticator{
		CRTokenFilename: craMockCRTokenFile,
		IAMProfileName:  craMockIAMProfileName,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Verify that RequestToken() returns a response with a valid refresh token.
	tokenResponse, err := auth.RequestToken()
	assert.Nil(t, err)
	assert.NotNil(t, tokenResponse)
	assert.Equal(t, craTestRefreshToken, tokenResponse.RefreshToken)
}

func TestCraAuthenticateSuccess(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	// Set up the authenticator to use the mock server for the IMDS and IAM operations.
	auth := &ComputeResourceAuthenticator{
		IAMProfileID:               craMockIAMProfileID,
		InstanceMetadataServiceURL: server.URL,
		URL:                        server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Create a new Request object to simulate an API request that needs authentication.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://myservice.localhost/api/v1", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Try to authenticate the request.
	err = auth.Authenticate(request)

	// Verify that it succeeded.
	assert.Nil(t, err)
	authHeader := request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+craTestAccessToken1, authHeader)

	// Call Authenticate again to make sure we used the cached access token.
	auth.Scope = "send-second-token"
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader = request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+craTestAccessToken1, authHeader)

	// Force expiration and verify that Authenticate() fetched the second access token.
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader = request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+craTestAccessToken2, authHeader)
}

func TestCraAuthenticateFailNoCRToken(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	// Set up the authenticator with both a bogus cr token filename and IMDS endpoint
	// so that we can't successfully retrieve a CR Token value.
	auth := &ComputeResourceAuthenticator{
		CRTokenFilename:            "bogus-cr-token-file",
		InstanceMetadataServiceURL: "http://bogus.imds.endpoint",
		IAMProfileName:             craMockIAMProfileName,
		URL:                        "https://bogus.iam.endpoint",
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Create a new Request object to simulate an API request that needs authentication.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://myservice.localhost/api/v1", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Try to authenticate the request (should fail)
	err = auth.Authenticate(request)

	// Validate the resulting error is a valid
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authErr)
	assert.EqualValues(t, authErr, err)
	// The casted error should match the original error message
	assert.Equal(t, err.Error(), authErr.Error())
}

func TestCraAuthenticateFailIAM(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	// Setup the authenticator to get the CR token from our mock server,
	// and set scope to cause the mock server to send a bad status code for the IAM call.
	auth := &ComputeResourceAuthenticator{
		InstanceMetadataServiceURL: server.URL,
		IAMProfileName:             craMockIAMProfileName,
		URL:                        server.URL,
		Scope:                      "status-bad-request",
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Create a new Request object to simulate an API request that needs authentication.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://myservice.localhost/api/v1", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Try to authenticate the request (should fail)
	err = auth.Authenticate(request)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authErr)
	assert.EqualValues(t, authErr, err)
	// The casted error should match the original error message
	assert.Contains(t, authErr.Error(), "Sorry, bad request!")
	assert.Equal(t, http.StatusBadRequest, authErr.Response.StatusCode)
}

func TestCraBackgroundTokenRefreshSuccess(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	auth := &ComputeResourceAuthenticator{
		CRTokenFilename: craMockCRTokenFile,
		IAMProfileName:  craMockIAMProfileName,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)

	// Now simulate being in the refresh window where the token is not expired but still needs to be refreshed.
	auth.getTokenData().RefreshTime = GetCurrentTime() - 1

	// Authenticator should detect the need to get a new access token in the background but use the current
	// cached access token for this next GetToken() call.
	// Set "scope" to cause the mock server to return the second access token the next time
	// we call the IAM "get token" operation.
	auth.Scope = "send-second-token"
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)

	// Wait for the background thread to finish.
	time.Sleep(2 * time.Second)
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken2, accessToken)
}

func TestCraBackgroundTokenRefreshFail(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	auth := &ComputeResourceAuthenticator{
		CRTokenFilename: craMockCRTokenFile,
		IAMProfileName:  craMockIAMProfileName,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)

	// Now simulate being in the refresh window where the token is not expired but still needs to be refreshed.
	auth.getTokenData().RefreshTime = GetCurrentTime() - 1

	// Authenticator should detect the need to get a new access token in the background but use the current
	// cached access token for this next GetToken() call.
	// Set "scope" to cause the mock server to return an error the next time the IAM "get token" operation is called.
	auth.Scope = "status-unauthorized"
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)

	// Wait for the background thread to finish.
	time.Sleep(2 * time.Second)

	// The background token refresh triggered by the previous GetToken() call above failed,
	// but the authenticator is still holding a valid, unexpired access token,
	// so this next GetToken() call should succeed and return the first access token
	// that we had previously cached.
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)

	// Next, simulate the expiration of the token, then we should expect
	// an error from GetToken().
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	accessToken, err = auth.GetToken()
	assert.NotNil(t, err)
	assert.Empty(t, accessToken)
	t.Logf("Expected error: %s\n", err.Error())
}

func TestCraClientIdAndSecret(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	auth := &ComputeResourceAuthenticator{
		CRTokenFilename: craMockCRTokenFile,
		IAMProfileName:  craMockIAMProfileName,
		ClientID:        craMockClientID,
		ClientSecret:    craMockClientSecret,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	auth.Scope = "check-basic-auth"
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)
}

func TestCraDisableSSL(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	auth := &ComputeResourceAuthenticator{
		CRTokenFilename:        craMockCRTokenFile,
		IAMProfileName:         craMockIAMProfileName,
		URL:                    server.URL,
		DisableSSLVerification: true,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)

	// Next, verify that the authenticator's Client is configured correctly.
	assert.NotNil(t, auth.Client)
	assert.NotNil(t, auth.Client.Transport)
	transport, ok := auth.Client.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.NotNil(t, transport.TLSClientConfig)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestCraUserHeaders(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	headers := make(map[string]string)
	headers["User-Header-1"] = "Value-1"
	headers["Host"] = "iam.cloud.ibm.com"

	auth := &ComputeResourceAuthenticator{
		CRTokenFilename: craMockCRTokenFile,
		IAMProfileName:  craMockIAMProfileName,
		URL:             server.URL,
		Headers:         headers,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	auth.Scope = "check-user-headers"
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)
}

func TestCraGetTokenTimeout(t *testing.T) {
	GetLogger().SetLogLevel(craTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	auth := &ComputeResourceAuthenticator{
		CRTokenFilename: craMockCRTokenFile,
		IAMProfileName:  craMockIAMProfileName,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, craTestAccessToken1, accessToken)

	// Next, tell the mock server to sleep for a bit, force the expiration of the token,
	// and configure the client with a short timeout.
	auth.Scope = "sleep"
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	auth.Client.Timeout = time.Second * 2
	accessToken, err = auth.GetToken()
	assert.Empty(t, accessToken)
	assert.NotNil(t, err)
	assert.NotNil(t, err.Error())
	t.Logf("Expected error: %s\n", err.Error())
	_, ok := err.(*AuthenticationError)
	assert.True(t, ok)
}