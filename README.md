# import-repository

Repository to store metadata about imports.

## Config

Simply set these environment variables (default values in brackets):
*    SERVER_PORT: port to listen on (8080)
*    JWT_PUB_RSA: public RSA Key to validate JWTs. If not set, JWTs will not be validated ("")
*    FORCE_AUTH: whether to enforce authentication (true)
*    FORCE_USER: whether to enforce a user id in the JWT (true)
*    IMPORT_TYPE_TOPIC: kafka Topic to publish import types on (import-types)
*    PERMISSIONS_URL: URL of the [permission-search](https://github.com/SENERGY-Platform/permission-search) (http://permissionsearch:8080)
*    MONGO_URL: URL of the mongo db (mongodb://localhost:27017)
*    MONGO_TABLE: mongo db table to use (importrepository)
*    MONGO_IMPORT_TYPE_COLLECTION: mongo collection to use for import types (importtype)
*    MONGO_REPL_SET: whether the mongo db is running as replication set (true)
*    ZOOKEEPER_URL: Zookeeper to connect to (localhost:2181)
*    GROUP_ID: group id to used to subscribe to kafka (import-repository)
*    VALIDATE: whether to validate import types of HTTP requests (false)
*    DEBUG: whether to print debug output (true)

## Data model

### ContentVariable
```
{
  "name": string,  
  "type": string,  
  "characteristic_id": string,  
  "sub_content_variables": ContentVariable[],
  "use_as_tag": bool
}
```

### ImportConfig
```
{
  "name": string,
  "description": string,
  "type": string,
  "default_value": any
}
```

### ImportType
```
{
  "id": string,
  "name": string,
  "description": string,
  "image": string,
  "default_restart": bool,
  "configs": ImportConfig[],
  "aspect_ids": string[],
  "output": ContentVariable,
  "function_ids": string[],
  "owner": string
}
```

## API

### Create
```
POST /device-types
Body: ImportType without id and owner (set automatically)
```

### Read
```
GET /device-types/:id
Returns the full ImportType
```

### Update
```
PUT /device-types/:id
Body: Full ImportType. Ensure id in url and ImportType match. Changing the owner is not allowed.
```

### Delete
```
DELETE /device-types/:id
```

## Security
Identity is provided by populating the Header "Authorization" with a JWT (prefixed by "Bearer ").
The token can be validated by providing a public RSA key as config.
Read/Write access is managed by [permission-command](https://github.com/SENERGY-Platform/permission-command)
and checked at [permission-search](https://github.com/SENERGY-Platform/permission-search).

## Validation
If enabled, all write requests will be validated.
* Type fields are checked for valid values:
    * https://schema.org/Text
    * https://schema.org/Integer
    * https://schema.org/Float
    * https://schema.org/Boolean
    * https://schema.org/ItemList
    * https://schema.org/StructuredValue
* References to characteristics, functions and aspects will be checked for existence at permission-search
* Configs may not have duplicate names
* Default values of configs must be of correct type
* Some fields may not be empty:
    * name
    * type
    * image
