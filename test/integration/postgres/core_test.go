//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/youenchene/EnglishCraft/internal/curriculum/adapters/content"
	"github.com/youenchene/EnglishCraft/internal/identity/adapters/password"
	identitypg "github.com/youenchene/EnglishCraft/internal/identity/adapters/postgres"
	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
	"github.com/youenchene/EnglishCraft/internal/learning/adapters/httpapi"
	learningpg "github.com/youenchene/EnglishCraft/internal/learning/adapters/postgres"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
	platformpg "github.com/youenchene/EnglishCraft/internal/platform/postgres"
)

// Each new core integration fixture owns a disposable schema; no application
// account or progression is altered, and cleanup only removes that schema.
func isolatedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not configured")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := "ec_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		if _, err := admin.Exec(ctx, `DROP SCHEMA `+schema+` CASCADE`); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	if err := platformpg.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return pool
}
func TestCoreTransactionsConcurrentRetriesAndReload(t *testing.T) {
	pool := isolatedPool(t)
	ctx := context.Background()
	clock := fixedClock{now: time.Now().UTC()}
	sessions := identity.NewSessions(identitypg.New(pool), password.Bcrypt{}, clock)
	if err := sessions.Provision(ctx, "parent", "parent-secret-for-test", "alex", "child-secret-for-test"); err != nil {
		t.Fatal(err)
	}
	catalog, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	store := learningpg.NewStore(pool)
	service := core.New(catalog, store, clock)
	if _, err := service.Start(ctx, "family-child", "lost-diamond"); err != nil {
		t.Fatal(err)
	}
	command := core.Submit{Student: "family-child", MissionID: "lost-diamond", Step: 0, ExerciseID: "diamond-go", Answer: "went", Key: "concurrent-retry-0001"}
	errs := make(chan error, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := service.Submit(ctx, command); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	state, err := store.Read(ctx, "family-child")
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Attempts) != 1 || state.Missions["lost-diamond"].XP != 10 {
		t.Fatal("concurrent retry duplicated evidence or reward")
	}
	reloaded := core.New(catalog, learningpg.NewStore(pool), clock)
	progress, err := reloaded.Progress(ctx, "family-child")
	if err != nil || progress.XP != 10 {
		t.Fatal("progress not retained after service restart")
	}
	command.Answer = "goed"
	if _, err := reloaded.Submit(ctx, command); !errors.Is(err, core.ErrConflict) {
		t.Fatal("changed retry payload accepted")
	}
	// Two different keys for the same current step are not a replay. Only one
	// submission can advance; the other must be rejected as stale.
	different := make(chan error, 2)
	for _, key := range []string{"different-key-0001", "different-key-0002"} {
		go func(key string) {
			_, err := reloaded.Submit(ctx, core.Submit{Student: "family-child", MissionID: "lost-diamond", Step: 1,
				ExerciseID: "diamond-found", Answer: "found", Key: key})
			different <- err
		}(key)
	}
	successes, conflicts := 0, 0
	for range 2 {
		err := <-different
		if err == nil {
			successes++
		} else if errors.Is(err, core.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatal("same step rewarded twice")
	}
	state, err = store.Read(ctx, "family-child")
	if err != nil || len(state.Attempts) != 2 || state.Missions["lost-diamond"].XP != 20 {
		t.Fatal("concurrent distinct-key progression invalid")
	}
}

// A failing callback must not commit any part of the state.
func TestCoreRollbackAndAuthPersistence(t *testing.T) {
	pool := isolatedPool(t)
	ctx := context.Background()
	clock := fixedClock{now: time.Now().UTC()}
	authStore := identitypg.New(pool)
	sessions := identity.NewSessions(authStore, password.Bcrypt{}, clock)
	if err := sessions.Provision(ctx, "parent", "parent-secret-for-test", "alex", "child-secret-for-test"); err != nil {
		t.Fatal(err)
	}
	store := learningpg.NewStore(pool)
	err := store.Transact(ctx, "family-child", func(s *core.StudentState) error { s.Receipts["test"] = core.Receipt{}; return errors.New("abort") })
	if err == nil {
		t.Fatal("expected abort")
	}
	state, err := store.Read(ctx, "family-child")
	if err != nil || len(state.Receipts) != 0 {
		t.Fatal("partial transaction committed")
	}
	token, _, err := sessions.Login(ctx, "alex", "child-secret-for-test")
	if err != nil {
		t.Fatal(err)
	}
	reloaded := identity.NewSessions(identitypg.New(pool), password.Bcrypt{}, clock)
	if p, err := reloaded.Resolve(ctx, token); err != nil || p.Role != identity.RoleChild {
		t.Fatal("session not persisted")
	}
	child, err := reloaded.Child(ctx, "family-parent")
	if err != nil || child != "family-child" {
		t.Fatal("parent ownership query failed")
	}
	if err := reloaded.Logout(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := reloaded.Resolve(ctx, token); err == nil {
		t.Fatal("revoked session persisted")
	}
	if err := sessions.Provision(ctx, "another", "another-parent-secret", "other-child", "another-child-secret"); err == nil {
		t.Fatal("initial accounts overwritten")
	}
}

func TestParentCreatesMultipleChildrenWithSeparateState(t *testing.T) {
	pool := isolatedPool(t)
	ctx := context.Background()
	clock := fixedClock{now: time.Now().UTC()}
	sessions := identity.NewSessions(identitypg.New(pool), password.Bcrypt{}, clock)
	if err := sessions.Provision(ctx, "parent", "parent-secret-for-test", "alex", "child-secret-for-test"); err != nil {
		t.Fatal(err)
	}
	parent := identity.Principal{SubjectID: "family-parent", Role: identity.RoleParent}
	catalog, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	httpapi.NewCore(core.New(catalog, learningpg.NewStore(pool), clock), sessions, false).Register(mux)
	handler := httpapi.SameOrigin(mux)
	parentToken, _, err := sessions.Login(ctx, "parent", "parent-secret-for-test")
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, path, body, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			r.AddCookie(&http.Cookie{Name: "englishcraft_session", Value: token})
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	if got := request("POST", "/api/v1/parent/children", `{"username":"anonymous","password":"long-password"}`, "").Code; got != 401 {
		t.Fatal("anonymous creation allowed", got)
	}
	child, err := sessions.AddChild(ctx, parent, "second-child", "second-child-password")
	if err != nil {
		t.Fatal(err)
	}
	children, err := sessions.Children(ctx, parent)
	if err != nil || len(children) != 2 {
		t.Fatal("children missing", err)
	}
	if got := request("GET", "/api/v1/parent/results?child="+child.ID, "", parentToken).Code; got != 200 {
		t.Fatal("new child results unavailable", got)
	}
	if got := request("GET", "/api/v1/parent/results?child=foreign", "", parentToken).Code; got != 403 {
		t.Fatal("foreign child results accepted", got)
	}
	created := request("POST", "/api/v1/parent/children", `{"username":"third-child","password":"third-child-password"}`, parentToken)
	if created.Code != 201 || strings.Contains(created.Body.String(), "password") {
		t.Fatal("HTTP creation failed or exposed password", created.Code)
	}
	_, principal, err := sessions.Login(ctx, "second-child", "second-child-password")
	if err != nil || principal.SubjectID != child.ID || principal.Role != identity.RoleChild {
		t.Fatal("new child cannot log in", err)
	}
	state, err := learningpg.NewStore(pool).Read(ctx, child.ID)
	if err != nil || len(state.Attempts) != 0 || len(state.Missions) != 0 {
		t.Fatal("new state not empty", err)
	}
	if _, err := sessions.AddChild(ctx, parent, "second-child", "second-child-password"); !errors.Is(err, identity.ErrUsernameTaken) {
		t.Fatal("duplicate accepted", err)
	}
	children, err = sessions.Children(ctx, parent)
	if err != nil || len(children) != 3 {
		t.Fatal("duplicate partially persisted")
	}
	if _, err := sessions.SelectChild(ctx, identity.Principal{SubjectID: "other-parent", Role: identity.RoleParent}, child.ID); !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Fatal("ownership bypass")
	}
	if err := identity.NewDeletion(sessions, identitypg.New(pool)).DeleteSelected(ctx, parent, child.ID); err != nil {
		t.Fatal(err)
	}
	children, err = sessions.Children(ctx, parent)
	if err != nil || len(children) != 2 || children[0].ID != "family-child" {
		t.Fatal("wrong child deleted")
	}
}
