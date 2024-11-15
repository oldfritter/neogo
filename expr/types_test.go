package expr

import (
	"github.com/oldfritter/neogo/db"
	"github.com/oldfritter/neogo/internal"
	"github.com/oldfritter/neogo/internal/tests"
)

func ExampleIdentifier_nil() {
	Match(
		db.Node(nil).To("e", nil),
	).Print()
	// Output:
	// MATCH ()-[e]->()
}

func ExampleIdentifier_string() {
	Match(
		db.Node("n"),
	).
		With("n").
		Print()
	// Output:
	// MATCH (n)
	// WITH n
}

func ExampleIdentifier_expr() {
	With(db.Qual(
		"timestamp()", // identifier
		"t",
	)).
		Print()
	// Output:
	// WITH timestamp() AS t
}

func ExampleIdentifier_pointer() {
	var p any

	With(db.Qual(&p, "pName")).
		Return(db.Qual(&p, "pDiffName")).
		Print()
	// Output:
	// WITH pName
	// RETURN pName AS pDiffName
}

func ExampleIdentifier_parameter() {
	data := []string{"a", "b", "c"}
	var n any

	Unwind(db.NamedParam(&data, "var"), "n").
		With(db.Qual(&n, "n")).
		Return(&n).
		Print()
	// Output:
	// UNWIND $var AS n
	// WITH n
	// RETURN n
}

func ExampleIdentifier_pointerToField() {
	var older, younger tests.Person

	Match(
		db.Node(
			db.Qual(&older, "older"),
		).
			To(
				tests.Knows{},
				db.Qual(&younger, "younger"),
			),
	).
		Where(db.Cond(&older.Age, ">", &younger.Age)).
		Return(&older.Name, &younger.Name).
		Print()
	// Output:
	// MATCH (older:Person)-[:KNOWS]->(younger:Person)
	// WHERE older.age > younger.age
	// RETURN older.name, younger.name
}

func ExampleMatch() {
	var m tests.Movie

	Match(
		db.Node(db.Var(
			tests.Person{},
			db.Props{
				"name": "'Oliver Stone'",
			},
		)).To(nil, db.Var("movie")),
	).
		Return(db.Qual(
			&m.Title,
			"movie.title",
		)).Print()
	// Output:
	// MATCH (:Person {name: 'Oliver Stone'})-->(movie)
	// RETURN movie.title
}

func ExampleOptionalMatch() {
	a := tests.Person{}
	r := tests.Directed{}

	Match(
		db.Node(db.Qual(
			&a, "a",
			db.Props{
				"name": "'Martin Sheen'",
			},
		)),
	).
		OptionalMatch(
			db.Node(&a).To(db.Qual(&r, "r"), nil),
		).Return(&a.Name, &r).Print()
	// Output:
	// MATCH (a:Person {name: 'Martin Sheen'})
	// OPTIONAL MATCH (a)-[r:DIRECTED]->()
	// RETURN a.name, r
}

func ExampleReturn() {
	var p tests.Person

	Match(db.Node(db.Qual(&p, "p", db.Props{"name": "'Keanu Reeves'"}))).
		Return(db.Qual(&p.Nationality, "citizenship")).Print()
	// Output:
	// MATCH (p:Person {name: 'Keanu Reeves'})
	// RETURN p.nationality AS citizenship
}

func ExampleWith() {
	var names []string

	Match(
		db.Node(db.Var("n", db.Props{"name": "'Anders'"})).
			Related(nil, "m"),
	).
		With(
			db.With("m", db.OrderBy("name", false), db.Limit("1")),
		).
		Match(db.Node("m").Related(nil, "o")).
		Return(db.Qual(names, "o.name")).Print()

	// Output:
	// MATCH (n {name: 'Anders'})--(m)
	// WITH m
	// ORDER BY m.name DESC
	// LIMIT 1
	// MATCH (m)--(o)
	// RETURN o.name
}

func ExampleSubquery() {
	var (
		p       tests.Person
		numConn int
	)

	Match(db.Node(db.Qual(&p, "p"))).
		Subquery(func(c *Client) Runner {
			return c.With(&p).
				Match(db.Node(&p).Related(nil, db.Var("c"))).
				Return(
					db.Qual(&numConn, "count(c)", db.Name("numberOfConnections")),
				)
		}).
		Return(&p.Name, &numConn).
		Print()

	// Output:
	// MATCH (p:Person)
	// CALL {
	//   WITH p
	//   MATCH (p)--(c)
	//   RETURN count(c) AS numberOfConnections
	// }
	// RETURN p.name, numberOfConnections
}

func ExampleCall() {
	var labels []string

	Call("db.labels()").
		Yield(db.Qual(&labels, "label")).
		Return(&labels).
		Print()

	// Output:
	// CALL db.labels()
	// YIELD label
	// RETURN label
}

func ExampleShow() {
	var (
		name any
		sig  string
	)

	Show("PROCEDURES").
		Yield(
			db.Qual(&name, "name"),
			db.Qual(&sig, "signature"),
		).
		Where(db.Cond(&name, "=", "'dbms.listConfig'")).
		Return(&sig).
		Print()

	// Output:
	// SHOW PROCEDURES
	// YIELD name, signature
	// WHERE name = 'dbms.listConfig'
	// RETURN signature
}

func ExampleUnwind() {
	events := map[string]any{
		"events": []map[string]any{
			{
				"id":   1,
				"year": 2014,
			},
			{
				"id":   2,
				"year": 2015,
			},
		},
	}
	type Year struct {
		internal.Node `neo4j:"Year"`

		Year int `json:"year"`
	}
	type Event struct {
		internal.Node `neo4j:"Event"`

		ID   int `json:"id"`
		Year int `json:"year"`
	}
	type In struct {
		internal.Relationship `neo4j:"IN"`
	}
	var (
		y Year
		e Event
	)

	Unwind(db.Qual(&events, "events"), "event").
		Merge(
			db.Node(db.Qual(&y, "y", db.Props{"year": "event.year"})),
		).
		Merge(
			db.Node(&y).
				From(In{}, db.Qual(&e, "e", db.Props{"id": "event.id"})),
		).
		Return(db.Return(db.Qual(&e.ID, "x"), db.OrderBy("", true))).
		Print()

	// Output:
	// UNWIND $events AS event
	// MERGE (y:Year {year: event.year})
	// MERGE (y)<-[:IN]-(e:Event {id: event.id})
	// RETURN e.id AS x
	// ORDER BY x
}

func ExampleCypher() {
	var n any

	Match(db.Node(db.Qual(&n, "n"))).
		Cypher(`WHERE n.name = 'Bob'`).
		Return(&n).
		Print()

	// Output:
	// MATCH (n)
	// WHERE n.name = 'Bob'
	// RETURN n
}

func ExampleUse() {
	var n any

	Use("myDatabase").
		Match(db.Node(db.Qual(&n, "n"))).
		Return("n").
		Print()
	// Output:
	// USE myDatabase
	// MATCH (n)
	// RETURN n
}

func ExampleUnion() {
	var name string
	Union(
		func(c *Client) Runner {
			return c.
				Match(db.Node(db.Var("n", db.Label("Person")))).
				Return(db.Qual(&name, "n.name", db.Name("name")))
		},
		func(c *Client) Runner {
			return c.
				Match(db.Node(db.Var("n", db.Label("Movie")))).
				Return(db.Qual(&name, "n.title", db.Name("name")))
		},
	).Print()

	// Output:
	// MATCH (n:Person)
	// RETURN n.name AS name
	// UNION
	// MATCH (n:Movie)
	// RETURN n.title AS name
}

func ExampleUnionAll() {
	var name string
	UnionAll(
		func(c *Client) Runner {
			return c.
				Match(db.Node(db.Var("n", db.Label("Person")))).
				Return(db.Qual(&name, "n.name", db.Name("name")))
		},
		func(c *Client) Runner {
			return c.
				Match(db.Node(db.Var("n", db.Label("Movie")))).
				Return(db.Qual(&name, "n.title", db.Name("name")))
		},
	).Print()

	// Output:
	// MATCH (n:Person)
	// RETURN n.name AS name
	// UNION ALL
	// MATCH (n:Movie)
	// RETURN n.title AS name
}

func ExampleYield() {
	var labels []string

	Call("db.labels()").
		Yield(db.Qual(&labels, "label")).
		Return(&labels).
		Print()

	// Output:
	// CALL db.labels()
	// YIELD label
	// RETURN label
}

func ExampleCreate() {
	var p any

	Create(db.Path(
		db.Node(db.Var(tests.Person{}, db.Props{"name": "'Andy'"})).
			To(tests.WorksAt{}, db.Var(tests.Company{}, db.Props{"name": "'Neo4j'"})).
			From(tests.WorksAt{}, db.Var(tests.Person{}, db.Props{"name": "'Michael'"})),
		"p",
	)).
		Return(db.Qual(&p, "p")).
		Print()

	// Output:
	// CREATE p = (:Person {name: 'Andy'})-[:WORKS_AT]->(:Company {name: 'Neo4j'})<-[:WORKS_AT]-(:Person {name: 'Michael'})
	// RETURN p
}

func ExampleMerge() {
	var person tests.Person

	Merge(
		db.Node(db.Qual(&person, "person")),
		db.OnMatch(
			db.SetPropValue(&person.Found, true),
			db.SetPropValue(&person.LastSeen, "timestamp()"),
		),
	).
		Return(&person.Name, &person.Found, &person.LastSeen).
		Print()

	// Output:
	// MERGE (person:Person)
	// ON MATCH
	//   SET
	//     person.found = true,
	//     person.lastSeen = timestamp()
	// RETURN person.name, person.found, person.lastSeen
}

func ExampleDelete() {
	var (
		n tests.Person
		r tests.ActedIn
	)

	Match(
		db.Node(db.Qual(&n, "n", db.Props{"name": "'Laurence Fishburne'"})).
			To(db.Qual(&r, "r"), nil),
	).
		Delete(&r).
		Print()

	// Output:
	// MATCH (n:Person {name: 'Laurence Fishburne'})-[r:ACTED_IN]->()
	// DELETE r
}

func ExampleDetachDelete() {
	var n tests.Person

	Match(
		db.Node(
			db.Qual(&n, "n",
				db.Props{"name": "'Carrie-Anne Moss'"},
			),
		),
	).
		DetachDelete(&n).
		Print()

	// Output:
	// MATCH (n:Person {name: 'Carrie-Anne Moss'})
	// DETACH DELETE n
}

func ExampleSet() {
	var n tests.Person

	Match(
		db.Node(db.Qual(&n, "n", db.Props{"name": "'Andy'"})),
	).
		Set(
			db.SetPropValue(&n.Position, "'Developer'"),
			db.SetPropValue(&n.Surname, "'Taylor'"),
		).
		Print()

	// Output:
	// MATCH (n:Person {name: 'Andy'})
	// SET
	//   n.position = 'Developer',
	//   n.surname = 'Taylor'
}

func ExampleRemove() {
	var n tests.Person
	var labels []string

	Match(db.Node(db.Qual(&n, "n", db.Props{"name": "'Peter'"}))).
		Remove(db.RemoveLabels(&n, "German", "Swedish")).
		Return(&n.Name, db.Qual(&labels, "labels(n)")).
		Print()

	// Output:
	// MATCH (n:Person {name: 'Peter'})
	// REMOVE n:German:Swedish
	// RETURN n.name, labels(n)
}

func ExampleForEach() {
	Match(
		db.Path(db.Node("start").To(db.Var(nil, db.VarLength("*")), "finish"), "p"),
	).
		Where(db.And(
			db.Cond("start.name", "=", "'A'"),
			db.Cond("finish.name", "=", "'D'"),
		)).
		ForEach("n", "nodes(p)", func(c *Updater[any, any]) {
			c.Set(db.SetPropValue("n.marked", true))
		}).
		Print()

	// Output:
	// MATCH p = (start)-[*]->(finish)
	// WHERE start.name = 'A' AND finish.name = 'D'
	// FOREACH (n IN nodes(p) | SET n.marked = true)
}

func ExampleWhere() {
	var n tests.Person

	Match(db.Node(db.Qual(&n, "n"))).
		Where(
			db.Or(
				db.Xor(
					db.Cond(&n.Name, "=", "'Peter'"),
					db.And(
						db.Cond(&n.Age, "<", "30"),
						db.Cond(&n.Name, "=", "'Timothy'"),
					),
				),
				db.Not(db.Or(
					db.Cond(&n.Name, "=", "'Timothy'"),
					db.Cond(&n.Name, "=", "'Peter'"),
				)),
			),
		).
		Return(
			db.Return(db.Qual(&n.Name, "name"), db.OrderBy("", true)),
			db.Qual(&n.Age, "age"),
		).Print()

	// Output:
	// MATCH (n:Person)
	// WHERE (n.name = 'Peter' XOR (n.age < 30 AND n.name = 'Timothy')) OR NOT (n.name = 'Timothy' OR n.name = 'Peter')
	// RETURN n.name AS name, n.age AS age
	// ORDER BY name
}
