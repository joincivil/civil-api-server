# GraphQL Development

## Relevant Files

`schema.graphql` - GraphQL Schema

`gqlgen.yml` - Configuration for `gqlgen`

`resolver.go` - Main resolver file

## Schema Naming Conventions

### Queries/Mutations
Queries and mutations follow GraphQL naming conventions ([GraphQL Specs](https://facebook.github.io/graphql/June2018/#sec-Schema)), but we prefix with an additional namespace. The query/mutation names are lower camel case. Namespace names should be one word and short, yet descriptive.

```
Spec:
<namespace><query/mutation field>

Examples:
type Query {
	tcrListings
	tcrGovernanceEvents
	newsroomArticles
	userCurrentUser
}

type Mutation {
	kycCreateApplicant
	userCreateUser
}
```

### Namespaces
Namespaces designate a group of related types, interfaces, and operations.  Each namespace must have a corresponding file containing all related resolvers for the namespace.  This will keep all related code together rather than in one giant sprawling file. 

```
Exs.
/pkg/graphql/resolver_crawl.go
/pkg/graphql/resolver_users.go
/pkg/graphql/resolver_kyc.go

```

`resolvers.go` will remain and contain the base resolvers.

### Special words
Words like `id`, `url`, `json` must be in all-caps unless it is it's on it's own.  ex. `userJSON`, `applicationID`, `serviceURL`, `id`, `uid`.


### schema.graphql organization

Namespaced types and interfaces must be grouped together in alphabetical order both in the file and within the `Query` and `Mutation` schemas.

```
Ex.

type Query {

	// Queries for User in alphabetical order
	...
 
	// Queries for KYC in alphabetical order
	...
}

type Mutation {

	// Mutations for User in alphabetical order
	...

	// Mutations for KYC in alphabetical order
	...
}

// Types/interfaces for User in alphabetical order
...

// Types/kyc for KYC in alphabetical order
...

// All scalars in alphabetical order
...

```

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
5. Implement code in `resolver.go` or `resolver_*.go` to populate data into the models.

## To Generate Code
```
make install-gorunpkg
cd pkg/graphql
gorunpkg github.com/99designs/gqlgen -v
```

## `resolver*.go`

`resolver.go` does not update on subsequent code generation via `gqlgen`, so it will need to be updated by hand. Will see if the `gqlgen` project will fix this in any future releases.

The options here are:

1. If it is small updates to fields or field types, it is straightforward to update by hand.
2. If there are new models or new endpoints, temporarily move the existing `resolver.go` and call the code generator again to generate a new stubbed out `resolver.go`.  Go through the new one and copy over the new stubbed resolvers over to the original `resolver.go`.  Implement the stubbed out resolvers in the original `resolver.go`

Run `make test` and/or `make lint` and check for errors and if things are not matching up.

## `dataloaden`
Reference: https://gqlgen.com/reference/dataloaders/ and https://github.com/vektah/dataloaden
 
To generate a dataloader file (i.e. for a `model.Listing` that creates `listingloader_gen.go`):
```
go get -u github.com/vektah/dataloaden
cd pkg/graphql
dataloaden -keys string github.com/joincivil/civil-events-processor/pkg/model.Listing
dataloaden -keys int github.com/joincivil/civil-events-processor/pkg/model.GovernanceEvent
```
Then implement code in `dataloaders.go` and modify `resolvers.go`
