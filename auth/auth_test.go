package auth_test

import (
	"context"
	"github.com/gabrielseibel1/gaef/auth"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticator_GetAuthenticatedUserID(t *testing.T) {
	canceledContext, cancel := context.WithCancel(context.TODO())
	cancel()

	type args struct {
		ctx   context.Context
		token string
	}
	tests := []struct {
		name    string
		server  *httptest.Server
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h := r.Header.Get("Authorization")
				if h != "Bearer test-token" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{\"id\":\"123\"}"))
			})),
			args: args{
				ctx:   context.TODO(),
				token: "test-token",
			},
			want:    "123",
			wantErr: false,
		},
		{
			name: "error message",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{\"error\":\"unauthorized\"}"))
			})),
			args: args{
				ctx: context.TODO(),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "header unauthorized",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("{\"id\":\"123\"}"))
			})),
			args: args{
				ctx: context.TODO(),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "malformed response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("malformed-response"))
			})),
			args: args{
				ctx: context.TODO(),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "canceled request context",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{\"id\":\"123\"}"))
			})),
			args: args{
				ctx: canceledContext,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.server.Close()
			a := auth.New(tt.server.URL)
			got, err := a.GetAuthenticatedUserID(tt.args.ctx, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuthenticatedUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetAuthenticatedUserID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
