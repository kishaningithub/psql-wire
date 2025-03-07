package wire

import (
	"context"
	"strconv"
	"testing"

	"github.com/lib/pq/oid"
	"github.com/stretchr/testify/assert"
)

func TestInvalidOptions(t *testing.T) {
	tests := [][]OptionFn{
		{
			Parse(func(context.Context, string) (PreparedStatementFn, []oid.Oid, error) { return nil, nil, nil }),
			SimpleQuery(func(context.Context, string, DataWriter, []string) error { return nil }),
		},
	}

	for index, test := range tests {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			srv := &Server{}
			for _, option := range test {
				err := option(srv)
				if err != nil {
					return
				}
			}

			t.Error("unexpected pass")
		})
	}
}

func TestSimpleQueryParameters(t *testing.T) {
	type test struct {
		query      string
		parameters []oid.Oid
	}

	tests := map[string]test{
		"positional": {
			query:      "SELECT * FROM users WHERE id = $1 AND age > $2",
			parameters: []oid.Oid{0, 0},
		},
		"unpositional": {
			query:      "SELECT * FROM users WHERE id = ? AND age > ?",
			parameters: []oid.Oid{0, 0},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			option := SimpleQuery(nil)

			srv := &Server{}
			err := option(srv)
			assert.NoError(t, err)

			statement, parameters, err := srv.Parse(context.Background(), test.query)
			assert.NoError(t, err)
			assert.NotNil(t, statement)
			assert.Equal(t, test.parameters, parameters)
		})
	}
}

func TestNilSessionHandler(t *testing.T) {
	srv, err := NewServer()
	assert.NoError(t, err)
	assert.NotNil(t, srv)

	bg := context.Background()
	ctx, err := srv.Session(bg)
	assert.NoError(t, err)
	assert.Equal(t, bg, ctx)
}

func TestSessionHandler(t *testing.T) {
	t.Parallel()

	type test []OptionFn

	type key string
	mock := key("key")
	value := "Super Secret Session ID"

	tests := map[string]test{
		"single": {
			Session(func(ctx context.Context) (context.Context, error) {
				return context.WithValue(ctx, mock, value), nil
			}),
		},
		"nested": {
			Session(func(ctx context.Context) (context.Context, error) {
				return ctx, nil
			}),
			Session(func(ctx context.Context) (context.Context, error) {
				return context.WithValue(ctx, mock, value), nil
			}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			srv, err := NewServer(test...)
			assert.NoError(t, err)
			assert.NotNil(t, srv)

			ctx, err := srv.Session(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, ctx)

			result := ctx.Value(mock)
			assert.Equal(t, value, result)
		})
	}
}
