package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CarbonFactory/chi"
)

func TestStripSlashes(t *testing.T) {
	r := chi.NewRouter()

	// This middleware must be mounted at the top level of the router, not at the end-handler
	// because then it'll be too late and will end up in a 404
	r.Use(StripSlashes)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("nothing here"))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	})

	r.Route("/accounts/{accountID}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			accountID := chi.URLParam(r, "accountID")
			w.Write([]byte(accountID))
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, resp := testRequest(t, ts, "GET", "/", nil); resp != "root" {
		t.Fatalf(resp)
	}
	if _, resp := testRequest(t, ts, "GET", "//", nil); resp != "root" {
		t.Fatalf(resp)
	}
	if _, resp := testRequest(t, ts, "GET", "/accounts/admin", nil); resp != "admin" {
		t.Fatalf(resp)
	}
	if _, resp := testRequest(t, ts, "GET", "/accounts/admin/", nil); resp != "admin" {
		t.Fatalf(resp)
	}
	if _, resp := testRequest(t, ts, "GET", "/nothing-here", nil); resp != "nothing here" {
		t.Fatalf(resp)
	}
}

func TestStripSlashesInRoute(t *testing.T) {
	r := chi.NewRouter()

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("nothing here"))
	})

	r.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	r.Route("/accounts/{accountID}", func(r chi.Router) {
		r.Use(StripSlashes)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("accounts index"))
		})
		r.Get("/query", func(w http.ResponseWriter, r *http.Request) {
			accountID := chi.URLParam(r, "accountID")
			w.Write([]byte(accountID))
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, resp := testRequest(t, ts, "GET", "/hi", nil); resp != "hi" {
		t.Fatalf(resp)
	}
	if _, resp := testRequest(t, ts, "GET", "/hi/", nil); resp != "nothing here" {
		t.Fatalf(resp)
	}
	if _, resp := testRequest(t, ts, "GET", "/accounts/admin", nil); resp != "accounts index" {
		t.Fatalf(resp)
	}
	if _, resp := testRequest(t, ts, "GET", "/accounts/admin/", nil); resp != "accounts index" {
		t.Fatalf(resp)
	}
	if _, resp := testRequest(t, ts, "GET", "/accounts/admin/query", nil); resp != "admin" {
		t.Fatalf(resp)
	}
	if _, resp := testRequest(t, ts, "GET", "/accounts/admin/query/", nil); resp != "admin" {
		t.Fatalf(resp)
	}
}

func TestRedirectSlashes(t *testing.T) {
	r := chi.NewRouter()

	// This middleware must be mounted at the top level of the router, not at the end-handler
	// because then it'll be too late and will end up in a 404
	r.Use(RedirectSlashes)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("nothing here"))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	})

	r.Route("/accounts/{accountID}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			accountID := chi.URLParam(r, "accountID")
			w.Write([]byte(accountID))
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	if resp, body := testRequest(t, ts, "GET", "/", nil); body != "root" && resp.StatusCode != 200 {
		t.Fatalf(body)
	}

	// NOTE: the testRequest client will follow the redirection..
	if resp, body := testRequest(t, ts, "GET", "//", nil); body != "root" && resp.StatusCode != 200 {
		t.Fatalf(body)
	}

	if resp, body := testRequest(t, ts, "GET", "/accounts/admin", nil); body != "admin" && resp.StatusCode != 200 {
		t.Fatalf(body)
	}

	// NOTE: the testRequest client will follow the redirection..
	if resp, body := testRequest(t, ts, "GET", "/accounts/admin/", nil); body != "admin" && resp.StatusCode != 200 {
		t.Fatalf(body)
	}

	if resp, body := testRequest(t, ts, "GET", "/nothing-here", nil); body != "nothing here" && resp.StatusCode != 200 {
		t.Fatalf(body)
	}

	// Ensure redirect Location url is correct
	{
		resp, body := testRequestNoRedirect(t, ts, "GET", "/accounts/someuser/", nil)
		if resp.StatusCode != 301 {
			t.Fatalf(body)
		}
		if resp.Header.Get("Location") != "/accounts/someuser" {
			t.Fatalf("invalid redirection, should be /accounts/someuser")
		}
	}

	// Ensure query params are kept in tact upon redirecting a slash
	{
		resp, body := testRequestNoRedirect(t, ts, "GET", "/accounts/someuser/?a=1&b=2", nil)
		if resp.StatusCode != 301 {
			t.Fatalf(body)
		}
		if resp.Header.Get("Location") != "/accounts/someuser?a=1&b=2" {
			t.Fatalf("invalid redirection, should be /accounts/someuser?a=1&b=2")
		}

	}
}
