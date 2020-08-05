# WSO-Go
The new flagship back-end for WSO's services. The WSO backend rewrite proposal is found [here](https://github.com/WilliamsStudentsOnline/wso-on-rails/wiki/Proposal:-WSO-Backend-Rewrite).

## Docs

## Running Locally

To run the server, simply do `make run-dev` or `make && ./wso-backend --development`.

Note: you must include a secrets file. So, run `cp config/secrets_example.yaml config/secrets.yaml` and edit the fields from there. You can also just set the environment variable `WSO_SECRET_JWT_SECRET_KEY=wso-jwt-development-secret`, which will work.

### Current Go Version: 1.14
It is worth noting that you should install Go via the official site, not a package repository like apt-get or brew, which often have outdated versions. You can find info on how to install Go [here](https://golang.org/doc/install).

## Onboarding 

### Learning Go
There are a number of resources out there to learn Go. The official tutorial is found [here](tour.golang.org). However, I prefer [Learn Go in Y Minutes](https://learnxinyminutes.com/docs/go/), which is pretty short and informative. 

### IDE
You can use whatever you want as your Go IDE. Personally, I use Intellij Goland, which you can get for free as a student. If you want something more lightweight, I suggest using Emacs.

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
- `jobs/` kubernetes job launching code and specific jobs to run on the server (e.g. update users from LDAP)
  - `dorms_update/data` dorm and dorm room data
- `k8s/` kubernetes configuration files
  - `base/` the base kubernetes configuration inherited by every deployment
  - `development/` the local development configuration
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

## Local Kubernetes Deployment
This is a guide to how to set up and run a local kubernetes deployment. Usually if you are just working on the API, 
it is okay to run the backend locally with `make` or `make run-dev`. But, if you need to make changes to the 
infrastructure, or you want to run the backend as if it was on production, this is your best bet. Please note that 
wso-dev can also function as a place to test your code in a kubernetes environment.

### Requirements

You will need [minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/#installation) as your local Kubernetes installation.
With minikube, you need a virtual machine hypervisor. I suggest [HyperKit](https://github.com/moby/hyperkit) for macOS. For other operating systems,
go with what looks best, but VirtualBox is always a staple.

You will also need to install [Docker](https://docs.docker.com/install/).

### Setting up

These instructions are for setting up for your first time.

Start up your minikube instance with:
```shell script
minikube start --vm-driver=VM-DRIVER-HERE
```

You also probably want to switch your docker client to use minikube, rather than its own installation. To do that, shut 
down your local docker machine if it is online. Then run:
```shell script
eval $(minikube docker-env)
```

You will need to do that every time you change shells or restart minikube.

Now, you can build the WSO-backend docker images:
```shell script
make docker-build-dev
```

Once they are built, you can deploy the kubernetes cluster with:
```shell script
make k8-apply-dev
```

To see the status of your cluster, open a new terminal window and run:
```shell script
minikube dashboard
```

If your wso-backend pod is in a failing loop (more than 3 fails). This is may be because the MySQL image does not 
contain the `development` database yet. If this is the case, run this:
```shell script
kubectl run -i --rm --image=mysql:8.0.17 --restart=Never mysql-client -- mysql -h mysql -ppassword < echo "create database development character set utf8mb4 collate utf8mb4_bin; exit;"
```
This command essentially deploys a MySQL kubernetes pod that connects to the database and creates the database. You may 
want to run `make k8-apply-dev` again after this to restart the wso-backend deployment

### Running

These instructions are for everyday running.

Tell docker to use minikube:
```shell script
eval $(minikube docker-env)
```

Build the docker images:
```shell script
make docker-build-dev
```

Open a dashboard:
```shell script
minikube dashboard
```

Deploy the kubernetes cluster:
```shell script
make k8-apply-dev
```
The last line of of this command outputs `Backend Service IP:`. This is the IP and port you use to connect to the 
backend service, rather than `http://localhost:8080`. Using that IP, you can now connect to the backend.

Take down the kubernetes cluster:
```shell script
make k8-delete-dev
```

To restart or reload the kubernetes cluster with a new image, just run `k8-delete-dev` followed by `k8-apply-dev`.

To access the MySQL database directly, execute this command, which will open up a terminal interface:
```shell script
kubectl run -n development -it --rm --image=mysql:8.0.17 --restart=Never mysql-client -- mysql -h mysql -ppassword development
```

To import existing SQL into the MySQL database, run this:
```shell script
kubectl run -i --rm --image=mysql:8.0.17 --restart=Never mysql-client -- mysql -h mysql -ppassword development < PATH_TO_SQL_DUMP_HERE
```

To set the default namespace to development:
```shell script
kubectl config set-context --current --namespace=development
```

To remove intermediate docker builds:
```shell script
docker image prune --filter label=stage=intermediate
```

Good luck!
