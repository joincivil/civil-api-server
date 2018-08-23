# GraphQL Development

`schema.graphql` - GraphQL Schema

`gqlgen.yml` - Configuration for `gqlgen`

`resolver.go` - Resolver code

## `gqlgen`

Read these before proceeding to update or add endpoints.

[Github](https://github.com/99designs/gqlgen)

[Docs](http://gqlgen.com)

[Examples](https://github.com/99designs/gqlgen/tree/master/example)


## To Update/Add To GraphQL

1. Update models, queries, inputs, etc. in `schema.graphql`
2. Write any models you want to use with the schema and put them into `pkg/models` (or let them autogenerate into `pkg/generated/graphql`).
3. Update `gqlgen.yml` with the locations of the models you have written.
4. Run the code generator as described below. This will generate `models.go` and `exec.go` and put them into `pkg/generated/graphql`	 
5. Implement code in `resolver.go` to populate data into the models.

## To Generate Code
```
make install-gorunpkg
cd pkg/graphql
gorunpkg github.com/99designs/gqlgen -v
```

## `resolver.go`

`resolver.go` does not update on subsequent code generation via `gqlgen`, so it will need to be updated by hand. Will see if the `gqlgen` project will fix this in any future releases.

The options here are:

1. If it is small updates to fields or field types, it is straightforward to update by hand.
2. If there are new models or new endpoints, temporarily move the existing `resolver.go` and call the code generator again to generate a new stubbed out `resolver.go`.  Go through the new one and copy over the new stubbed resolvers over to the original `resolver.go`.  Implement the stubbed out resolvers in the original `resolver.go`

Run `make test` and/or `make lint` and check for errors and if things are not matching up.


