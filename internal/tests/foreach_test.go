package tests

import (
	"testing"

	"github.com/oldfritter/neogo/db"
	"github.com/oldfritter/neogo/internal"
)

// TODO: probs needs more tests lol
func TestForEach(t *testing.T) {
	t.Run("Return a limited subset of the rows", func(t *testing.T) {
		c := internal.NewCypherClient()
		cy, err := c.
			Match(
				db.Path(db.Node("start").To(db.Var(nil, db.VarLength("*")), "finish"), "p"),
			).
			Where(db.And(
				db.Cond("start.name", "=", "'A'"),
				db.Cond("finish.name", "=", "'D'"),
			)).
			ForEach("n", "nodes(p)", func(c *internal.CypherUpdater[any]) {
				c.Set(db.SetPropValue("n.marked", true))
			}).
			Compile()

		Check(t, cy, err, internal.CompiledCypher{
			Cypher: `
					MATCH p = (start)-[*]->(finish)
					WHERE start.name = 'A' AND finish.name = 'D'
					FOREACH (n IN nodes(p) | SET n.marked = true)
					`,
		})
	})
}
