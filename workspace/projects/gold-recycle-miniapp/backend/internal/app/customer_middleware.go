/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: customer_middleware.go
 * 功能描述: 顾客端认证中间件
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

package app

import (
	"context"
	"net/http"
	"strings"
)

const (
	customerKey        contextKey = "customer"
	customerSessionKey contextKey = "customer_session"
)

// withCustomerAuth 校验顾客 token，顾客 token 不可通过 withAuth，员工 token 不可通过 withCustomerAuth。
func (a *App) withCustomerAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" || token == r.Header.Get("Authorization") {
			a.writeError(w, r, http.StatusUnauthorized, 40101, "missing bearer token")
			return
		}

		session, profile, ok := a.store.getCustomerByToken(token)
		if !ok {
			a.writeError(w, r, http.StatusUnauthorized, 40102, "invalid or expired customer token")
			return
		}

		if profile.Status != "active" {
			a.writeError(w, r, http.StatusForbidden, 40302, "customer account is disabled")
			return
		}

		ctx := context.WithValue(r.Context(), customerKey, profile)
		ctx = context.WithValue(ctx, customerSessionKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func currentCustomer(ctx context.Context) (CustomerProfile, bool) {
	customer, ok := ctx.Value(customerKey).(CustomerProfile)
	return customer, ok
}

func mustCurrentCustomer(ctx context.Context) CustomerProfile {
	customer, _ := ctx.Value(customerKey).(CustomerProfile)
	return customer
}

func currentCustomerSession(ctx context.Context) (CustomerSession, bool) {
	session, ok := ctx.Value(customerSessionKey).(CustomerSession)
	return session, ok
}
