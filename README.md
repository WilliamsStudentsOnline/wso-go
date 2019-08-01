# WSO-Go
The new flagship back-end for WSO's services. The WSO backend rewrite proposal is found [here](https://github.com/WilliamsStudentsOnline/wso-on-rails/wiki/Proposal:-WSO-Backend-Rewrite).

## Running

To run the server, simply do `make run-dev` or `./wso-backend --development`.

Note: you must include a secrets file. So, run `cp config/secrets_example.yaml config/secrets.yaml` and edit the fields from there. You can also just set the environment variable `WSO_SECRET_JWT_SECRET_KEY=wso-jwt-development-secret`, which will work.

### Current Go Version: 1.12
It is worth noting that you should install Go via the official site, not a package repository like apt-get or brew, which often have outdated versions. You can find info on how to install Go [here](https://golang.org/doc/install).

## Development

### Git Workflow/Pipeline
Steps for a 10/10 development workflow:

1. Notice an issue/feature and create a GitHub issue.
2. Checkout a feature branch: the name should be `feature/YOUR-FEATURE-HERE` or `feature/YOUR-NAME/YOUR-FEATURE-HERE`.
3. Write the code and create the tests.
    * Please follow this [helpful guide](https://github.com/golang/go/wiki/CodeReviewComments) on how to write commit-worthy Go code.
4. Make a pull request and link your original issue.
5. After approval merge the pull request by squashing all of your commits into one.

### Services
This project uses microservices to define API endpoints. This is essentially the combination of a controller and a router. Look at the user service for a good example.

### Models
The models folder will contain all models. Currently, there are two types of structs for each model. Help with the database driver can be found [here](https://gorm.io/docs).
The schema struct (e.g. `user_schema.go` or `User{}`) is the parsed Go interpretation of a database row. Any functions built off the schema struct, should relate directly to the data at hand (such as `IsStudent()`), and not make any DB calls.
The model struct (e.g. `user.go` or `UserModel{}`) is the database adapter for this model. It should contain the DB as a field and will run any CRUD or other DB-related queries. Most of these queries should return a schema struct.

#### Schema
Note that in the schema is defined following the [GORM guidelines](https://gorm.io/docs/models). Optional fields are pointer-type, and associations are documented [here](https://gorm.io/docs/belongs_to.html). When working with any optional fields, you can easily convert a literal value into a pointer by using the `lib/to_pointer.go` file, which has functions like `lib.StrToPtr(str string) *string`.

### REST-API Guidelines
* Use plural names for resources (when nouns): e.g. use `/users`, rather than `/user`.
* When resources are verbs or adjectives, use whatever fits best.
* Use dashes when resources must be more than one word: e.g. use `/areas-of-study`, rather than `/area_of_study` or `/areaOfStudy`.

### Auto-Generate
You can use the auto-generator to generate a services and models. Usage is as follows:

For Services:
`go run lib/generate/cmd/main.go service -m [model] [service_name]`

For Models:
`go run lib/generate/cmd/main.go model -t [table] [ModelName] `

### Lib
The library (lib) folder contains tools that multiple other folders and files use. No file in the lib folder should import any code from another place in this repo (external places are fine though).

### Migrations
To generate a database migration, run the command `go run db/migrations/cmd/main.go -m <ModelName> -t <table_name> <migration_title>`

### Config
The config folder contains all of the configuration & secrets parsers. It also sets up the database and does necessary middleware.

### Building
To build the Go binary, run `go build -tags=jsoniter -o wso-go main.go`. You can then just execute `./wso-go`.

## API Endpoints
Get All Users:
```http request
GET localhost:8080/api/v1/users
```
Get User:
```http request
GET localhost:8080/api/v1/users/:user_id
```
Update User:
```http request
PUT localhost:8080/api/v1/users/:user_id
{
    "visible": true,
    "dorm_visible": true,
    "home_visible": true,
    "pronoun": "",
    "off_cycle": false
}
```
Authenticate/Login:
```http request
POST localhost:8080/api/v1/auth/login
{
    "unix_id": "admin",
    "password": "doesnt matter"
}
```
Refresh JWT Token:
```http request
GET localhost:8080/api/v1/auth/refresh_token
```

### Authentication Flow
We use something called a [JWT](jwt.io), or JSON Web Token for the API. This allows us to keep sessions and verify user identities without cookies or database queries. It works like this:
1. A user will request a token from the `auth/login` endpoint. They will pass in their login credentials, which will be checked with LDAP (not implemented yet).
1. If the user is verified, the server will then pull their user from the DB and create a payload. This payload will consist of the user's ID and the scopes the user is allowed (e.g. if the user is a senior, they can go to ephcatch; if the user is an admin, they can do other queries; if the user is not signed in but on school wifi, they can be read only).
1. The server will then take this payload and sign it with its secret key, before handing the JWT back to the user.
1. The user now can add the header `Authorization: Bearer <JWT GOES HERE>` to any request and be authenticated and allowed to access other API endpoints (like `user`)
1. The JWT has a one hour timeout (we can change this). After an hour, the JWT becomes invalid and the user must sign in again.
1. Alternatively, before the hour is up, a user can query the `auth/refresh_token` endpoint to get a new token without having to sign in again.

## Structure

- `config/` contains configurations for server
  - `auth_middleware.go` is the JWT authentication middleware that runs on _all_ API calls. Learn more about the authentication process above
  - `config.go` loads the environment config files into go
  - `database.go` loads the database connection/configuration
  - `scope_middleware.go` is the scope-based role authorization package running on specific API calls. Learn more above
  - `environment/*.yml` are configuration yaml files named by the environment it is run in
- `services/` contains all server microservices
  - `base.go` every service should inherit useful methods from base. But, base should only have external imports.
  - `*/` other service folders with defined parts
- `models/` contains database models
  - `base_model.go` every model inherits the base model; it contains important properties/funcs for all models
  - `department_schema.go` is the database schema for the departments table
  - `user.go` is the model for a user; it includes all database-side functions we run from controllers
  - `user_schema.go` is the database schema for the users table
- `lib/` library files (helpful functions, etc.)
- `test.db` the database of generated data the demo server uses
- `main.go` the entry-point of the code; contains all routing information

## Local Kubernetes Deployment
This is a guide to how to set up and run a local kubernetes deployment. Usually if you are just working on the API, 
it is okay to run the backend locally with `go build` and `go run`. But, if you need to make changes to the 
infrastructure, or you want to run the backend as if it was on production, this is your best bet. Please note that 
wso-dev can also function as a place to test your code in a kubernetes environment.

How-to guide coming soon.
`eval $(minikube docker-env)`
`create database development character set utf8mb4 collate utf8mb4_bin;`
`docker build -t wso-backend:dev-latest .`