// +build integration

package nrsignup_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"

	cemail "github.com/joincivil/go-common/pkg/email"
	ctime "github.com/joincivil/go-common/pkg/time"
)

func TestHandleGrantApprovalNotRequested(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	emailer := cemail.NewEmailerWithSandbox(sendGridKey, true)
	tokenGen := auth.NewJwtTokenGenerator([]byte(testSecret))
	userService := buildUserService()
	jsonbService := buildJsonbService(t)

	signupService, _ := nrsignup.NewNewsroomSignupService(
		emailer,
		userService,
		jsonbService,
		tokenGen,
		"http://localhost:8080",
	)

	tokenSub := fmt.Sprintf("1:%v", nrsignup.ApprovedSubValue)

	now := ctime.CurrentEpochSecsInInt()
	expires := 60

	token, _ := tokenGen.GenerateToken(tokenSub, now+expires)

	status, approved := nrsignup.HandleGrantApproval(
		token,
		tokenGen,
		signupService,
	)

	if status != nrsignup.GrantApprovalStatusError || approved {
		t.Errorf("Should have gotten an error: %v, %v", status, approved)
	}
}

func TestHandleGrantApprovalApproved(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	emailer := cemail.NewEmailerWithSandbox(sendGridKey, true)
	tokenGen := auth.NewJwtTokenGenerator([]byte(testSecret))
	userService := buildUserService()
	jsonbService := buildJsonbService(t)

	signupService, _ := nrsignup.NewNewsroomSignupService(
		emailer,
		userService,
		jsonbService,
		tokenGen,
		"http://localhost:8080",
	)

	tokenSub := fmt.Sprintf("1:%v", nrsignup.ApprovedSubValue)

	now := ctime.CurrentEpochSecsInInt()
	expires := 60

	token, _ := tokenGen.GenerateToken(tokenSub, now+expires)

	err := signupService.RequestGrant("1", buildTestNewsroom())
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	status, approved := nrsignup.HandleGrantApproval(
		token,
		tokenGen,
		signupService,
	)

	if status != nrsignup.GrantApprovalStatusOK || !approved {
		t.Errorf("Should have approved the grant: %v, %v", status, approved)
	}
}

func TestHandleGrantApprovalRejected(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	emailer := cemail.NewEmailerWithSandbox(sendGridKey, true)
	tokenGen := auth.NewJwtTokenGenerator([]byte(testSecret))
	userService := buildUserService()
	jsonbService := buildJsonbService(t)

	signupService, _ := nrsignup.NewNewsroomSignupService(
		emailer,
		userService,
		jsonbService,
		tokenGen,
		"http://localhost:8080",
	)

	tokenSub := fmt.Sprintf("1:%v", nrsignup.RejectedSubValue)

	now := ctime.CurrentEpochSecsInInt()
	expires := 60

	token, _ := tokenGen.GenerateToken(tokenSub, now+expires)

	err := signupService.RequestGrant("1", buildTestNewsroom())
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	status, approved := nrsignup.HandleGrantApproval(
		token,
		tokenGen,
		signupService,
	)

	if status != nrsignup.GrantApprovalStatusOK || approved {
		t.Errorf("Should have rejected the grant: %v, %v", status, approved)
	}
}

func TestHandleGrantApprovalInvalidToken(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	emailer := cemail.NewEmailerWithSandbox(sendGridKey, true)
	tokenGen := auth.NewJwtTokenGenerator([]byte(testSecret))
	userService := buildUserService()
	jsonbService := buildJsonbService(t)

	signupService, _ := nrsignup.NewNewsroomSignupService(
		emailer,
		userService,
		jsonbService,
		tokenGen,
		"http://localhost:8080",
	)

	err := signupService.RequestGrant("1", buildTestNewsroom())
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	token := "thisisabadtoken"
	status, approved := nrsignup.HandleGrantApproval(
		token,
		tokenGen,
		signupService,
	)

	if status != nrsignup.GrantApprovalStatusInvalidToken || approved {
		t.Errorf("Should have returned an invalid token error: %v, %v", status, approved)
	}
}

func TestHandleGrantApprovalNoApprovedValue(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	emailer := cemail.NewEmailerWithSandbox(sendGridKey, true)
	tokenGen := auth.NewJwtTokenGenerator([]byte(testSecret))
	userService := buildUserService()
	jsonbService := buildJsonbService(t)

	signupService, _ := nrsignup.NewNewsroomSignupService(
		emailer,
		userService,
		jsonbService,
		tokenGen,
		"http://localhost:8080",
	)

	// No delimiter for the token sub
	tokenSub := fmt.Sprintf("1%v", nrsignup.RejectedSubValue)

	now := ctime.CurrentEpochSecsInInt()
	expires := 60

	token, _ := tokenGen.GenerateToken(tokenSub, now+expires)

	err := signupService.RequestGrant("1", buildTestNewsroom())
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	status, approved := nrsignup.HandleGrantApproval(
		token,
		tokenGen,
		signupService,
	)

	if status != nrsignup.GrantApprovalStatusInvalidToken || approved {
		t.Errorf("Should have returned an invalid token error: %v, %v", status, approved)
	}
}

func TestHandleGrantApprovalBadApprovedValue(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	emailer := cemail.NewEmailerWithSandbox(sendGridKey, true)
	tokenGen := auth.NewJwtTokenGenerator([]byte(testSecret))
	userService := buildUserService()
	jsonbService := buildJsonbService(t)

	signupService, _ := nrsignup.NewNewsroomSignupService(
		emailer,
		userService,
		jsonbService,
		tokenGen,
		"http://localhost:8080",
	)

	// No delimiter for the token sub
	tokenSub := fmt.Sprintf("1:badvalue")

	now := ctime.CurrentEpochSecsInInt()
	expires := 60

	token, _ := tokenGen.GenerateToken(tokenSub, now+expires)

	err := signupService.RequestGrant("1", buildTestNewsroom())
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	status, approved := nrsignup.HandleGrantApproval(
		token,
		tokenGen,
		signupService,
	)

	if status != nrsignup.GrantApprovalStatusInvalidToken || approved {
		t.Errorf("Should have returned an invalid token error: %v, %v", status, approved)
	}
}

func TestApproveGrantHandler(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	emailer := cemail.NewEmailerWithSandbox(sendGridKey, true)
	tokenGen := auth.NewJwtTokenGenerator([]byte(testSecret))
	userService := buildUserService()
	jsonbService := buildJsonbService(t)

	signupService, _ := nrsignup.NewNewsroomSignupService(
		emailer,
		userService,
		jsonbService,
		tokenGen,
		"http://localhost:8080",
	)

	err := signupService.RequestGrant("1", buildTestNewsroom())
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	tokenSub := fmt.Sprintf("1:%v", nrsignup.ApprovedSubValue)

	now := ctime.CurrentEpochSecsInInt()
	expires := 60

	token, _ := tokenGen.GenerateToken(tokenSub, now+expires)

	requestURI := fmt.Sprintf("/nrsignup/grantapprove/%v", token)

	req, err := http.NewRequest("GET", requestURI, nil)
	if err != nil {
		t.Fatal(err)
	}

	config := &nrsignup.NewsroomSignupApproveGrantConfig{
		NrsignupService: signupService,
		TokenGenerator:  tokenGen,
	}

	httpRec := httptest.NewRecorder()
	handlerFunc := nrsignup.NewsroomSignupApproveGrantHandler(config)
	handler := http.HandlerFunc(handlerFunc)

	// Add it to the chi context so it works.
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(nrsignup.GrantApproveTokenURLParam, token)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.ServeHTTP(httpRec, req)

	status := httpRec.Code
	if status != http.StatusOK {
		t.Errorf("Should have returned OK response code: %v", status)
	}

	body := httpRec.Body.String()
	if !strings.Contains(body, "Thanks, it worked!") {
		t.Errorf("Handler should have responded with a success message")
	}
}

func TestApproveGrantHandlerInvalidToken(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	emailer := cemail.NewEmailerWithSandbox(sendGridKey, true)
	tokenGen := auth.NewJwtTokenGenerator([]byte(testSecret))
	userService := buildUserService()
	jsonbService := buildJsonbService(t)

	signupService, _ := nrsignup.NewNewsroomSignupService(
		emailer,
		userService,
		jsonbService,
		tokenGen,
		"http://localhost:8080",
	)

	err := signupService.RequestGrant("1", buildTestNewsroom())
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	tokenSub := fmt.Sprintf("badtokeninput")

	now := ctime.CurrentEpochSecsInInt()
	expires := 60

	token, _ := tokenGen.GenerateToken(tokenSub, now+expires)

	requestURI := fmt.Sprintf("/nrsignup/grantapprove/%v", token)

	req, err := http.NewRequest("GET", requestURI, nil)
	if err != nil {
		t.Fatal(err)
	}

	config := &nrsignup.NewsroomSignupApproveGrantConfig{
		NrsignupService: signupService,
		TokenGenerator:  tokenGen,
	}

	httpRec := httptest.NewRecorder()
	handlerFunc := nrsignup.NewsroomSignupApproveGrantHandler(config)
	handler := http.HandlerFunc(handlerFunc)

	// Add it to the chi context so it works.
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(nrsignup.GrantApproveTokenURLParam, token)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.ServeHTTP(httpRec, req)

	status := httpRec.Code
	if status != http.StatusOK {
		t.Errorf("Should have returned OK response code: %v", status)
	}

	body := httpRec.Body.String()
	if !strings.Contains(body, "token is invalid") {
		t.Errorf("Handler should have responded with an invalid message")
	}
}

func TestApproveGrantHandlerError(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	emailer := cemail.NewEmailerWithSandbox(sendGridKey, true)
	tokenGen := auth.NewJwtTokenGenerator([]byte(testSecret))
	userService := buildUserService()
	jsonbService := buildJsonbService(t)

	signupService, _ := nrsignup.NewNewsroomSignupService(
		emailer,
		userService,
		jsonbService,
		tokenGen,
		"http://localhost:8080",
	)

	tokenSub := fmt.Sprintf("1:%v", nrsignup.ApprovedSubValue)

	now := ctime.CurrentEpochSecsInInt()
	expires := 60

	token, _ := tokenGen.GenerateToken(tokenSub, now+expires)

	requestURI := fmt.Sprintf("/nrsignup/grantapprove/%v", token)

	req, err := http.NewRequest("GET", requestURI, nil)
	if err != nil {
		t.Fatal(err)
	}

	config := &nrsignup.NewsroomSignupApproveGrantConfig{
		NrsignupService: signupService,
		TokenGenerator:  tokenGen,
	}

	httpRec := httptest.NewRecorder()
	handlerFunc := nrsignup.NewsroomSignupApproveGrantHandler(config)
	handler := http.HandlerFunc(handlerFunc)

	// Add it to the chi context so it works.
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(nrsignup.GrantApproveTokenURLParam, token)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.ServeHTTP(httpRec, req)

	status := httpRec.Code
	if status != http.StatusOK {
		t.Errorf("Should have returned OK response code: %v", status)
	}

	body := httpRec.Body.String()
	if !strings.Contains(body, "internal problem") {
		t.Errorf("Handler should have responded with an invalid message")
	}
}
