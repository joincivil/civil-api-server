# JSON Store GraphQL Endpoint

GraphQL endpoint to store and retrieve arbitrary JSON blobs. Pass the JSON as a string and retrieve it as a JSON string and/or list of key/value pairs from the JSON.

## Schema
```
type Query {
  jsonb(id: String, hash: String): [Jsonb]!
}

input JsonbInput {
  id: String!
  jsonStr: String!
}

type Mutation {
  createJsonb(input: JsonbInput!): Jsonb!
}

type JsonField {
  key: String!
  value: JsonFieldValue!
}

type Jsonb {
  id: String!
  hash: String!
  createdDate: Time!
  rawJson: String!
  json: [JsonField!]!
}

```
## Sample Queries
```
# Mutation to add a JSONb value.  'id' is a given "group" identifier like a user id or session 
# id.
mutation {
  createJsonb(input:{
    id: <Some ID>
    jsonStr: <JSON str>
  }){
    id
    hash
    createdDate
    rawJson
    json {
      key
      value
    }
  }
}

# Query to retrieve all JSONb for a particular 'id'.
{
 jsonb(id:<Some ID>) {
    id
    hash
    createdDate
    rawJson
    json {
      key
      value
    }
  }
}

# Query to retrieve a specific JSONb by it's hash.
{
 jsonb(has:<JSONb Hash>) {
    id
    hash
    createdDate
    rawJson
    json {
      key
      value
    }
  }
}

```


## Persistence Info
There is an interface defined to allow us to store the data into any brand of DB `pkg/jsonstore/models.JsonbPersister`.  The persister is initiated and added to the GraphQL resolver in `main.go`.  Feel free to implement some other one.

Currently, there is a `Postgresql` implementation of the interface which is our default. The JSON blob is stored as a `jsonb` Postgres type.  This will allow us to run different queries on the JSON data if needed.
