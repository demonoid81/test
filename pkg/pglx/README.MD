## Querying

The handle types in pglx implement the same basic verbs for querying your database:

- `Exec(...) (pgconn.CommandTag, error)` - unchanged from pgx
- `Query(...) (*pgx.Rows, error)` - unchanged from pgx
- `QueryRow(...) *pgx.Row` - unchanged from pgx

These extensions to the built-in verbs:

- `MustExec() pgconn.CommandTag` -- Exec, but panic on error
- `Queryx(...) (*pglx.Rows, error)` - Query, but return an pglx.Rows
- `QueryRowx(...) *pglx.Row` -- QueryRow, but return an pglx.Row

And these new semantics:

- `Get(dest interface{}, ...) error`
- `Select(dest interface{}, ...) error`

Let's go from the unchanged interface through the new semantics, explaining their use.

### Exec ¶
Exec and MustExec get a connection from the connection pool and executes the provided query on the server. For drivers that do not support ad-hoc query execution, a prepared statement may be created behind the scenes to be executed. The connection is returned to the pool before the result is returned.

```bazaar
schema := 'CREATE TABLE place (
country text,
city text NULL,
telcode integer);'

// execute a query on the server
result, err := db.Exec(schema)

// or, you can use MustExec, which panics on error
cityState := 'INSERT INTO place (country, telcode) VALUES (?, ?)'
countryCity := 'INSERT INTO place (country, city, telcode) VALUES (?, ?, ?)'
db.MustExec(cityState, "Hong Kong", 852)
db.MustExec(cityState, "Singapore", 65)
db.MustExec(countryCity, "South Africa", "Johannesburg", 27)
```

The result return **pgconn.CommandTag**

###Query

Query is the primary way to run queries with pgx that return row results. Query returns an pgx.Rows object and an error:

// fetch all places from the db
rows, err := db.Query("SELECT country, city, telcode FROM place")

// iterate over each row
for rows.Next() {
var country string
// note that city can be NULL, so we use the NullString type
var city    sql.NullString
var telcode int
err = rows.Scan(&country, &city, &telcode)
}
// check the error from rows
err = rows.Err()
You should treat the Rows like a database cursor rather than a materialized list of results. Although driver buffering behavior can vary, iterating via Next() is a good way to bound the memory usage of large result sets, as you're only scanning a single row at a time. Scan() uses reflect to map sql column return types to Go types like string, []byte, et al. If you do not iterate over a whole rows result, be sure to call rows.Close() to return the connection back to the pool!

The error returned by Query is any error that might have happened while preparing or executing on the server. This can include grabbing a bad connection from the pool, although database/sql will retry 10 times to attempt to find or create a working connection. Generally, the error will be due to bad SQL syntax, type mismatches, or incorrect field and table names.

In most cases, Rows.Scan will copy the data it gets from the driver, as it is not aware of how the driver may reuse its buffers. The special type sql.RawBytes can be used to get a zero-copy slice of bytes from the actual data returned by the driver. After the next call to Next(), such a value is no longer valid, as that memory might have been overwritten by the driver.

The connection used by the Query remains active until either all rows are exhausted by the iteration via Next, or rows.Close() is called, at which point it is released. For more information, see the section on the connection pool.

The pglx extension Queryx behaves exactly as Query does, but returns an pglx.Rows, which has extended scanning behaviors:

type Place struct {
Country       string
City          sql.NullString
TelephoneCode int `db:"telcode"`
}

rows, err := db.Queryx("SELECT * FROM place")
for rows.Next() {
var p Place
err = rows.StructScan(&p)
}
The primary extension on pglx.Rows is StructScan(), which automatically scans results into struct fields. Note that the 
fields must be exported (capitalized) in order for pglx to be able to write into them, something true of all 
marshallers in Go. You can use the db struct tag to specify which column name maps to each struct field, or set a new 
default mapping with db.MapperFunc(). The default behavior is to use strings.Lower on the field name to match against 
the column names. For more information about StructScan, SliceScan, and MapScan, see the section on advanced scanning.

###QueryRow

QueryRow fetches one row from the server. It takes a connection from the connection pool and executes the query using 
Query, returning a Row object which has its own internal Rows object:

```bazaar
row := db.QueryRow("SELECT * FROM place WHERE telcode=?", 852)
var telcode int
err = row.Scan(&telcode)
```

Unlike Query, QueryRow returns a Row type result with no error, making it safe to chain the Scan off of the return. 
If there was an error executing the query, that error is returned by Scan. If there are no rows, Scan returns ErrNoRows. 
If the scan itself fails (eg. due to type mismatch), that error is also returned.

The Rows struct internal to the Row result is Closed upon Scan, meaning that the connection used by QueryRow is kept 
open until the result is scanned. It also means that sql.RawBytes is not usable here, since the referenced memory 
belongs to the driver and may already be invalid by the time control is returned to the caller.

The pglx extension QueryRowx will return an pglx.Row instead of an pgx.Row, and it implements the same scanning 
extensions as Rows, outlined above and in the advanced scanning section:

```bazaar
var p Place
err := db.QueryRowx("SELECT city, telcode FROM place LIMIT 1").StructScan(&p)
```


Get and Select
Get and Select are time saving extensions to the handle types. They combine the execution of a query with flexible scanning semantics. To explain them clearly, we have to talk about what it means to be scannable:

a value is scannable if it is not a struct, eg string, int
a value is scannable if it implements scannable
a value is scannable if it is a struct with no exported fields (eg. time.Time)
Get and Select use rows.Scan on scannable types and rows.StructScan on non-scannable types. They are roughly analagous to QueryRow and Query, where Get is useful for fetching a single result and scanning it, and Select is useful for fetching a slice of results:

```bazaar
p := Place{}
pp := []Place{}

// this will pull the first place directly into p
err = db.Get(&p, "SELECT * FROM place LIMIT 1")

// this will pull places with telcode > 50 into the slice pp
err = db.Select(&pp, "SELECT * FROM place WHERE telcode > ?", 50)

// they work with regular types as well
var id int
err = db.Get(&id, "SELECT count(*) FROM place")

// fetch at most 10 place names
var names []string
err = db.Select(&names, "SELECT name FROM place LIMIT 10")
```

Get and Select both will close the Rows they create during query execution, and will return any error encountered at any step of the process. Since they use StructScan internally, the details in the advanced scanning section also apply to Get and Select.

Select can save you a lot of typing, but beware! It's semantically different from Queryx, since it will load the entire result set into memory at once. If that set is not bounded by your query to some reasonable size, it might be best to use the classic Queryx/StructScan iteration instead.

###Transactions

To use transactions, you must create a transaction handle with DB.Begin(). Code like this will not work:

```bazaar
// this will not work if connection pool > 1
db.MustExec("BEGIN;")
db.MustExec(...)
db.MustExec("COMMIT;")
```

Remember, Exec and all other query verbs will ask the DB for a connection and then return it to the pool each time. There's no guarantee that you will receive the same connection that the BEGIN statement was executed on. To use transactions, you must therefore use DB.Begin()

```bazaar
tx, err := db.Begin()
err = tx.Exec(...)
err = tx.Commit()
```

The DB handle also has the extensions Beginx() and MustBegin(), which return an pglx.Tx instead of an pgx.Tx:

```bazaar
tx := db.MustBegin()
tx.MustExec(...)
err = tx.Commit()
```

pglx.Tx has all of the handle extensions that pglx.DB has.

Since transactions are connection state, the Tx object must bind and control a single connection from the pool. A Tx 
will maintain that single connection for its entire life cycle, releasing it only when Commit() or Rollback() is called. 
You should take care to call at least one of these, or else the connection will be held until garbage collection.

Because you only have one connection to use in a transaction, you can only execute one statement at a time; the cursor 
types Row and Rows must be Scanned or Closed, respectively, before executing another query. If you attempt to send the 
server data while it is sending you a result, it can potentially corrupt the connection.

Finally, Tx objects do not actually imply any behavior on the server; they merely execute a BEGIN statement and bind a 
single connection. The actual behavior of the transaction, including things like locking and isolation, is completely 
unspecified and database dependent.

##Query Helpers

The database/sql package does not do anything with your actual query text. This makes it trivial to use backend-specific features in your code; you can write queries just as you would write them in your database prompt. While this is very flexible, it makes writing certain kinds of queries difficult.

"In" Queries
Because database/sql does not inspect your query and it passes your arguments directly to the driver, it makes dealing with queries with IN clauses difficult:

SELECT * FROM users WHERE level IN (?);
When this gets prepared as a statement on the backend, the bindvar ? will only correspond to a single argument, but what is often desired is for that to be a variable number of arguments depending on the length of some slice, eg:

var levels = []int{4, 6, 7}
rows, err := db.Query("SELECT * FROM users WHERE level IN (?);", levels)
This pattern is possible by first processing the query with pglx.In:

var levels = []int{4, 6, 7}
query, args, err := pglx.In("SELECT * FROM users WHERE level IN (?);", levels)

// pglx.In returns queries with the `?` bindvar, we can rebind it for our backend
query = db.Rebind(query)
rows, err := db.Query(query, args...)
What pglx.In does is expand any bindvars in the query passed to it that correspond to a slice in the arguments to the length of that slice, and then append those slice elements to a new arglist. It does this with the ? bindvar only; you can use db.Rebind to get a query suitable for your backend.

###Named Queries

Named queries are common to many other database packages. They allow you to use a bindvar syntax which refers to the names of struct fields or map keys to bind variables a query, rather than having to refer to everything positionally. The struct field naming conventions follow that of StructScan, using the NameMapper and the db struct tag. There are two extra query verbs related to named queries:

NamedQuery(...) (*pglx.Rows, error) - like Queryx, but with named bindvars
NamedExec(...) (pgconn.CommandTag, error) - like Exec, but with named bindvars
And one extra handle type:

NamedStmt - an pglx.Stmt which can be prepared with named bindvars
// named query with a struct
p := Place{Country: "South Africa"}
rows, err := db.NamedQuery(`SELECT * FROM place WHERE country=:country`, p)

// named query with a map
m := map[string]interface{}{"city": "Johannesburg"}
result, err := db.NamedExec(`SELECT * FROM place WHERE city=:city`, m)
Named query execution and preparation works off both structs and maps. If you desire the full set of query verbs, prepare a named statement and use that instead:

p := Place{TelephoneCode: 50}
pp := []Place{}

// select all telcodes > 50
nstmt, err := db.PrepareNamed(`SELECT * FROM place WHERE telcode > :telcode`)
err = nstmt.Select(&pp, p)
Named query support is implemented by parsing the query for the :param syntax and replacing it with the bindvar supported by the underlying database, then performing the mapping at execution, so it is usable on any database that pglx supports. You can also use pglx.Named, which uses the ? bindvar, and can be composed with pglx.In:

arg := map[string]interface{}{
"published": true,
"authors": []{8, 19, 32, 44},
}
query, args, err := pglx.Named("SELECT * FROM articles WHERE published=:published AND author_id IN (:authors)", arg)
query, args, err := pglx.In(query, args...)
query = db.Rebind(query)
db.Query(query, args...)
Advanced Scanning
StructScan is deceptively sophisticated. It supports embedded structs, and assigns to fields using the same precedence rules that Go uses for embedded attribute and method access. A common use of this is sharing common parts of a table model among many tables, eg:

type AutoIncr struct {
ID       uint64
Created  time.Time
}

type Place struct {
Address string
AutoIncr
}

type Person struct {
Name string
AutoIncr
}
With the structs above, Person and Place will both be able to receive id and created columns from a StructScan, because they embed the AutoIncr struct which defines them. This feature can enable you to quickly create an ad-hoc table for joins. It works recursively as well; the following will have the Person's Name and its AutoIncr ID and Created fields accessible, both via the Go dot operator and via StructScan:

type Employee struct {
BossID uint64
EmployeeID uint64
Person
}
Note that pglx historically supported this feature for non-embedded structs, this ended up being confusing because users were using this feature to define relationships and embedding the same structs twice:

```bazaar
type Child struct {
    Father Person
    Mother Person
}
```

This causes some problems. In Go, it's legal to shadow descendent fields; if Employee from the embedded example defined a Name, it would take precedence over the Person's Name. But ambiguous selectors are illegal and cause a runtime error. If we wanted to create a quick JOIN type for Person and Place, where would we put the id column, which is defined in both via their embedded AutoIncr? Would there be an error?

Because of the way that pglx builds the mapping of field name to field address, by the time you Scan into a struct, it no longer knows whether or not a name was encountered twice during its traversal of the struct tree. So unlike Go, StructScan will choose the "first" field encountered which has that name. Since Go struct fields are ordered from top to bottom, and pglx does a breadth-first traversal in order to maintain precedence rules, it would happen in the shallowest, top-most definition. For example, in the type:

```bazaar
type PersonPlace struct {
    Person
    Place
}
```

A StructScan will set an id column result in Person.AutoIncr.ID, also accessible as Person.ID. To avoid confusion, it's suggested that you use AS to create column aliases in your SQL instead.

Scan Destination Safety
By default, StructScan will return an error if a column does not map to a field in the destination. This mimics the treatment for things like unused variables in Go, but does not match the way that standard library marshallers like encoding/json behave. Because SQL is generally executed in a more controlled fashion than parsing JSON, and these errors are generally coding errors, a decision was made to return errors by default.

Like unused variables, columns which you ignore are a waste of network and database resources, and detecting things 
like an incompatible mapping or a typo in a struct tag early can be difficult without the mapper letting you know 
something wasn't found.

Despite this, there are some cases where ignoring columns with no destination might be desired. For this, there is the 
Unsafe method on each Handle type which returns a new copy of that handle with this safety turned off:


```
var p Person
// err here is not nil because there are no field destinations for columns in 'place'
err = db.Get(&p, "SELECT * FROM person, place LIMIT 1;")

// this will NOT return an error, even though place columns have no destination
udb := db.Unsafe()
err = udb.Get(&p, "SELECT * FROM person, place LIMIT 1;")
```


Alternate Scan Types
In addition to using Scan and StructScan, an pglx Row or Rows can be used to automatically return a slice or a map of 
results:

```
rows, err := db.Queryx("SELECT * FROM place")
for rows.Next() {
// cols is an []interface{} of all of the column results
    cols, err := rows.SliceScan()
}

rows, err := db.Queryx("SELECT * FROM place")
for rows.Next() {
    results := make(map[string]interface{})
    err = rows.MapScan(results)
}
```

SliceScan returns an []interface{} of all columns, which can be useful in situations where you are executing queries on
behalf of a third party and have no way of knowing what columns may be returned. MapScan behaves the same way, but maps 
the column names to interface{} values. An important caveat here is that the results returned by **rows.Columns()** does 
not
include fully qualified names, such that SELECT a.id, b.id FROM a NATURAL JOIN b will result in a Columns result of 
[]string{"id", "id"}, clobbering one of the results in your map.