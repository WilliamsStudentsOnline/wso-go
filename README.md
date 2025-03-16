# WSO-Go
The new flagship back-end for WSO's services. The WSO backend rewrite proposal is found [here](https://github.com/WilliamsStudentsOnline/wso-on-rails/wiki/Proposal:-WSO-Backend-Rewrite).

## Docs

WSO-Go auto-generates API docs found on: 
 - The WSO-DEV server (you need a VPN or to be on campus) here: http://wso-dev.williams.edu/api/docs/index.html
 - Locally here: http://localhost:8080/docs/index.html

## Running Locally 

To run the server, simply do `make run-dev` or `make && ./wso-backend --development`.

Note: you must include a secrets file. So, run `cp config/secrets_example.yaml config/secrets.yaml` and edit the fields from there. You can also just set the environment variable `WSO_SECRET_JWT_SECRET_KEY=wso-jwt-development-secret`, which will work.

### Redis

**NOTE**: As of November 2024, wso-go uses Redis for several non-critical database tables, such as saved classes in the course scheduler. For most cases, this is not important, but if you intend to develop for these features, deploy Redis with:
-  `docker build -t wso-redis -f lib/redis_util/Redis.Dockerfile lib/redis_util`
- `docker run --name wso-redis-instance -d -p 6379:6379 wso-redis`

When you are finished, kill the container with:
- `docker kill wso-redis-instance`

To restart the container if you regret killing it:
- `docker restart wso-redis-instance`

If you want to remove the container:
- `docker remove wso-redis-instance`
- `docker rmi wso-redis`

All data is lost when a container is removed. For data persistence, use the flag `-v /absolute/path/on/your/computer/:data` on `docker run` to dump data to a path of your choice.

### Grafana and Prometheus

wso-go can be run with Grafana and Prometheus to improve the quality and accessiblity of logs. This feature is non-critical and you do not need to use it. To install Grafana and Prometheus, you should read [this](https://grafana.com/docs/grafana/latest/setup-grafana/installation/debian/) site for Grafana and [this](https://prometheus.io/docs/prometheus/latest/installation/) site for Prometheus. Open source version of both can be installed using a package manager like `apt-get` or `brew`. Please also install the package `prometheus-node-exporter`, or else many Grafana graphs will not work correctly.

Once you have these packages installed, you can optionally install the Grafana config files to your system Grafana directory by running `grafana-install` as root in the `prod_files/` directory. To run them alongside wso-go, run the command `make run-with-analytics`. This should autostart the servers. Grafana is accessible on port 9093 through a browser, and Prometheus is configured to run on port 9095. To login to Grafana for the first time, use the username `admin` and the password `admin`. All files will be found in the `prod_files/` directory. Please change the values in `run-analytics.sh` when you use this in production. 

Alternatively, if you're running wso-go on a Linux system, you can install the systemd service files and edit the values to run Grafana and Prometheus. This is how it is actually done in production.

### Current Go Version: 1.21
It is worth noting that you should install Go via the official site, not a package repository like `apt-get` or `brew`, which often have outdated versions. You can find info on how to install Go [here](https://golang.org/doc/install).

## Onboarding 

### Learning Go
There are a number of resources out there to learn Go. The official tutorial is found [here](tour.golang.org). However, I prefer [Learn Go in Y Minutes](https://learnxinyminutes.com/docs/go/), which is pretty short and informative. 

### IDE
You can use whatever you want as your Go IDE. VSCode is a popular choice as a lightweight editor. If you would like a full IDE, JetBrains Goland is free for students and it provides many more features.

### First Issue
Choose an unassigned issue tagged "good first issue" and reach out to the Backend team lead for more information and guidance. If you want some examples of good wso-go code, check out `wso-go/services/ephmatch` or `wso-go/services/users`.

## Development

### Git Workflow/Pipeline
Steps for a 10/10 development workflow:

1. Notice an issue/feature and create a GitHub issue.
2. Checkout a feature branch: the name should be `feature/YOUR-FEATURE-HERE` or `feature/YOUR-NAME/YOUR-FEATURE-HERE`.
3. Write the code and create the tests.
    * Please follow this [helpful guide](https://github.com/golang/go/wiki/CodeReviewComments) on how to write commit-worthy Go code.
4. Make a pull request and link your original issue.
    * The pull request will be automatically tested on Jenkins. If it passes, you can just ignore it. However, if Jenkins fails and you want to see why, *you must be on the Williams network to access Jenkins*.
5. After approval merge the pull request by squashing all of your commits into one.

**CHANGES INFO:** Before committing any changes, run `make commit` to autoformat and update your code.

### Services
This project uses microservices to define API endpoints. This is essentially the combination of a controller and a router. Look at the dormtrak service for a good example.
All controller endpoints **MUST BE DOCUMENTED** in the style laid out (swaggo).

### Models
The models folder will contain all models. Currently, there are two types of structs for each model. Help with the database driver can be found [here](https://gorm.io/docs).
The schema struct (e.g. `user_schema.go` or `User{}`) is the parsed Go interpretation of a database row. Any functions built off the schema struct, should relate directly to the data at hand (such as `IsStudent()`), and not make any DB calls.
The model struct (e.g. `user.go` or `UserModel{}`) is the database adapter for this model. It should contain the DB as a field and will run any CRUD or other DB-related queries. Most of these queries should return a schema struct.

#### Schema
Note that in the schema is defined following the [GORM guidelines](https://gorm.io/docs/models). Optional fields and all booleans are pointer-type, and associations are documented [here](https://gorm.io/docs/belongs_to.html). When working with any optional fields, you can easily convert a literal value into a pointer by using the `lib/to_pointer.go` file, which has functions like `lib.StrToPtr(str string) *string`.

### REST-API Guidelines
* Use plural names for resources (when nouns): e.g. use `/users`, rather than `/user`.
* When resources are verbs or adjectives, use whatever fits best.
* Use dashes when resources must be more than one word: e.g. use `/areas-of-study`, rather than `/area_of_study` or `/areaOfStudy` ([source](https://restfulapi.net/resource-naming/)).
* Array query parameters should be passed and named as with brackets: e.g. the struct field `preload []string` would become `?preload[]=foo&preload[]=bar`.

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
To build the Go binary, run `make`. You can then just execute `./wso-backend`.

## API Endpoints

API Endpoints are documented at `localhost:8080/docs`, and in the director `docs/` as swagger files. You can also 
look at controller comments for any endpoint info. Don't use the provided query tools, bc they don't play nice 
with our authentication.

### Grafana
Analytics are available at port `:9092`!

## Authentication Flow
*NOTE: THIS IS DEPRECATED*
We use something called a [JWT](jwt.io), or JSON Web Token for the API. This allows us to keep sessions and verify user identities without cookies or database queries. It works like this:
1. A user will request a token from the `auth/login` endpoint. They will pass in their login credentials, which will be checked with LDAP (not implemented yet).
2. If the user is verified, the server will then pull their user from the DB and create a payload. This payload will consist of the user's ID and the scopes the user is allowed (e.g. if the user is a senior, they can go to ephcatch; if the user is an admin, they can do other queries; if the user is not signed in but on school wifi, they can be read only).
3. The server will then take this payload and sign it with its secret key, before handing the JWT back to the user.
4. The user now can add the header `Authorization: Bearer <JWT GOES HERE>` to any request and be authenticated and allowed to access other API endpoints (like `user`)
5. The JWT has a timeout. After that timeout is over, the JWT becomes invalid and the user must sign in again.
6. Alternatively, before the timeout ends, a user can query the `auth/refresh-token` endpoint to get a new token without having to sign in again.
7. The `auth/update-token` endpoint also updates a new token without signing in, but this does a DB query and will assign new information to the payload.

## Structure

- `config/` contains the server configuration library, logging library, secrets library, and various configurations
  - `environment/development.yml` is the local development configuration yaml
- `db/` migration code (and dummy SQLite databases)
  - `migrations/` specific database migrations
- `docs/` swagger API docs to be compiled
- `jobs/` cron job launching code and specific jobs to run on the server (e.g. update users from LDAP)
  - `dorms_update/data` dorm and dorm room data
- `lib/` library files (helpful functions, errors, etc.). We try to minimize the number of external libraries we import here, as this is so widely used
  - `errors.go` contains all API errors
  - `auth/` contains authentication info about scopes and useful ways to use scopes
  - `autocomplete/` contains logic to for various model autocompletion
    - `mysql/` the MySQL backend for autocomplete
  - `generate/` contains command to generate services and models
  - `ldap/` contains LDAP interfacing code
  - `search/` contains logic for various model searching
  - `test_utils/` utilities for tests to use
- `models/` contains database models
- `server` master API routing and entrypoint for entire server
- `services/` contains all server microservices
- `Dockerfile.*` dockerfiles for various tasks and builds

To remove intermediate docker builds:
```shell script
docker image prune --filter label=stage=intermediate
```

Good luck!
