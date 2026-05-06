package main

// CHQ: Claude AI refactored this file
// middleware.go

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const (
    contextKeyIsAdmin  contextKey = "isAdmin"
    contextKeyUserID   contextKey = "userID"   // int (users.id)
    contextKeyDescopeID contextKey = "descopeID" // string (google-XXXXXX)
)

func requireAdminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        isAdmin, ok := r.Context().Value(contextKeyIsAdmin).(bool)
        if !ok || !isAdmin {
            writeJSONError(w, http.StatusForbidden, "admin access required")
            return
        }
        next.ServeHTTP(w, r)
    })
}

func sessionValidationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Extract Bearer token
        sessionToken := r.Header.Get("Authorization")
        if sessionToken == "" {
            writeJSONError(w, http.StatusUnauthorized, "no session token provided")
            return
        }
        sessionToken = strings.TrimPrefix(sessionToken, "Bearer ")

        ctx := r.Context()

        // 2. Validate with Descope
        authorized, token, err := descopeClient.Auth.ValidateSessionWithToken(ctx, sessionToken)
        if err != nil || !authorized {
            log.Printf("session validation failed: %v", err)
            writeJSONError(w, http.StatusUnauthorized, "invalid session token")
            return
        }

        // 3. Extract Descope ID
        descopeID := token.ID
        if descopeID == "" {
            writeJSONError(w, http.StatusUnauthorized, "descope ID missing from token")
            return
        }

        // 4. Check admin role
        isAdmin := descopeClient.Auth.ValidateRoles(ctx, token, []string{"Game Admin"})

        // 5. Resolve or create user
        userID, err := resolveOrCreateUser(ctx, descopeID)
        if err != nil {
            log.Printf("user resolution failed: %v", err)
            writeJSONError(w, http.StatusInternalServerError, "internal server error")
            return
        }

        // 6. Store in context
        ctx = context.WithValue(ctx, contextKeyDescopeID, descopeID)
        ctx = context.WithValue(ctx, contextKeyUserID, userID)
        ctx = context.WithValue(ctx, contextKeyIsAdmin, isAdmin)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}