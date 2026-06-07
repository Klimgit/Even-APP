package swagger

import (
	"strings"
	"testing"
)

func TestSpecBuilderMergesBackends(t *testing.T) {
	sb := &specBuilder{}
	if err := sb.init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	authDoc, err := unmarshalSpec([]byte(`
openapi: 3.0.3
info:
  title: auth
  version: 0.0.0
paths:
  /api/v1/auth/login:
    post:
      operationId: login
      responses:
        "200":
          description: OK
components:
  schemas:
    User:
      type: object
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
`))
	if err != nil {
		t.Fatalf("auth spec: %v", err)
	}
	if err := sb.add(authDoc); err != nil {
		t.Fatalf("add auth: %v", err)
	}

	lexiconDoc, err := unmarshalSpec([]byte(`
openapi: 3.0.3
info:
  title: lexicon
  version: 0.0.0
paths:
  /api/v1/platform/media/presign:
    post:
      operationId: platformMediaPresign
      responses:
        "200":
          description: OK
components:
  schemas:
    ErrorResponse:
      type: object
  responses:
    DefaultError:
      description: err
`))
	if err != nil {
		t.Fatalf("lexicon spec: %v", err)
	}
	if err := sb.add(lexiconDoc); err != nil {
		t.Fatalf("add lexicon: %v", err)
	}

	out, err := sb.build().Render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"/api/v1/health:",
		"/api/v1/auth/login:",
		"/api/v1/platform/media/presign:",
		"ErrorResponse:",
		"bearerAuth:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in merged spec:\n%s", want, text)
		}
	}
}

func TestSpecBuilderRejectsDuplicatePaths(t *testing.T) {
	sb := &specBuilder{}
	if err := sb.init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	doc, err := unmarshalSpec([]byte(`
openapi: 3.0.3
info:
  title: svc
  version: 0.0.0
paths:
  /api/v1/auth/me:
    get:
      responses:
        "200":
          description: OK
`))
	if err != nil {
		t.Fatalf("spec: %v", err)
	}
	if err := sb.add(doc); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := sb.add(doc); err == nil {
		t.Fatal("expected duplicate path error")
	}
}
