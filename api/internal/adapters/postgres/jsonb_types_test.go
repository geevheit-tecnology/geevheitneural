package postgres_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
)

// Colunas jsonb precisam usar json.RawMessage. Com []byte, o pgx em protocolo simples
// (usado atrás do pooler do Supabase) envia bytea e o Postgres rejeita com
// "invalid input syntax for type json", abortando a transação.
func TestJSONBParamsUseRawMessage(t *testing.T) {
	want := reflect.TypeOf(json.RawMessage{})
	cases := map[string]any{
		"CreateCardEventParams.Payload": sqlcgen.CreateCardEventParams{}.Payload,
		"CardEvent.Payload":             sqlcgen.GeevCardEvent{}.Payload,
	}
	for name, v := range cases {
		if got := reflect.TypeOf(v); got != want {
			t.Errorf("%s: tipo %v, esperado json.RawMessage", name, got)
		}
	}
}
