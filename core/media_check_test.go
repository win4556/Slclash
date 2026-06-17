package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mediaCheckRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn mediaCheckRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func mediaCheckTextResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestCheckChatGPTAllowsAlternateReachableProbe(t *testing.T) {
	client := &http.Client{
		Transport: mediaCheckRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Host + req.URL.Path {
			case "chatgpt.com/cdn-cgi/trace":
				return mediaCheckTextResponse("loc=JP\n"), nil
			case "ios.chat.openai.com/":
				return nil, context.DeadlineExceeded
			case "api.openai.com/compliance/cookie_requirements":
				return mediaCheckTextResponse("{}"), nil
			default:
				return nil, errors.New("unexpected request: " + req.URL.String())
			}
		}),
	}

	result := checkChatGPT(context.Background(), client)
	if result.Status != "clean" {
		t.Fatalf("expected clean, got %+v", result)
	}
	if result.Region != "JP" {
		t.Fatalf("expected JP region, got %q", result.Region)
	}
}

func TestCheckYouTubeFallsBackToReachabilityProbe(t *testing.T) {
	client := &http.Client{
		Transport: mediaCheckRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Host + req.URL.Path {
			case "www.youtube.com/premium":
				return nil, context.DeadlineExceeded
			case "www.youtube.com/generate_204":
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			default:
				return nil, errors.New("unexpected request: " + req.URL.String())
			}
		}),
	}

	result := checkYouTube(context.Background(), client)
	if result.Status != "unknown" {
		t.Fatalf("expected unknown reachability fallback, got %+v", result)
	}
	if result.Evidence != "generate_204" {
		t.Fatalf("expected generate_204 evidence, got %q", result.Evidence)
	}
}

func TestCheckYouTubeUsesRegionAsAvailableSignal(t *testing.T) {
	client := &http.Client{
		Transport: mediaCheckRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host+req.URL.Path == "www.youtube.com/premium" {
				return mediaCheckTextResponse(`{"INNERTUBE_CONTEXT_GL":"JP"}`), nil
			}
			return nil, errors.New("unexpected request: " + req.URL.String())
		}),
	}

	result := checkYouTube(context.Background(), client)
	if result.Status != "available" {
		t.Fatalf("expected available, got %+v", result)
	}
	if result.Region != "JP" {
		t.Fatalf("expected JP region, got %q", result.Region)
	}
}
