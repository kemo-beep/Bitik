package rbac

import (
	"context"
	"fmt"

	rbacstore "github.com/bitik/backend/internal/store/rbac"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/jackc/pgx/v5/pgtype"
)

const modelText = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch2(r.obj, p.obj) && (r.act == p.act || p.act == "*")
`

func textString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// NewEnforcer builds a Casbin enforcer from rows in casbin_rule (ptype "p" only).
func NewEnforcer(ctx context.Context, q *rbacstore.Queries) (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(m, false)
	if err != nil {
		return nil, err
	}
	rules, err := q.ListCasbinRules(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rules {
		if row.Ptype != "p" {
			continue
		}
		sub := textString(row.V0)
		obj := textString(row.V1)
		act := textString(row.V2)
		if sub == "" || obj == "" || act == "" {
			continue
		}
		if _, err := e.AddPolicy(sub, obj, act); err != nil {
			return nil, fmt.Errorf("casbin add policy: %w", err)
		}
	}
	e.EnableAutoSave(false)
	return e, nil
}

// AllowAnyRole returns true if any role matches (subject in policy is role name).
func AllowAnyRole(e *casbin.Enforcer, roles []string, path, method string) bool {
	if e == nil {
		return false
	}
	for _, r := range roles {
		if ok, _ := e.Enforce(r, path, method); ok {
			return true
		}
	}
	return false
}
