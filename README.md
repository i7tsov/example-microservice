# Example microservice

This is an example microservice that iimplements HTTP API
for adding users to database and listing existing users.

## Building

Use `build.sh` script, it will set the correct version of the binary.

Otherwise standard `go build ./cmd/server` works as well for the developement purposes.

## Configuring

To find list of applicable config options, run

```sh
./bin/server --help
```

Service accepts all of command line flags, environment variables or config
file in YAML format.

Config file `config.yaml` is first checked at the current folder,
then in `/etc/<service-name>`.

Example config file is located in the root of repository.

Options from the environment variables are using capital case with underscores
instead of dots. 

Full example:

```sh
export SERVER_PORT=11111
./bin/server --log.level=debug
# other options are taken from config.yaml
```

## Running database

To run postgres database in container, first install `docker`.
Then, run:

```sh
docker run --name exampledb -e POSTGRES_PASSWORD=example -e POSTGRES_USER=example -e POSTGRES_DB=users -v ~/db/example:/var/lib/postgresql/data -d -p 5432:5432 postgres:16.3-alpine3.20
```

## Running service

Use:

```sh
./bin/server
```

Alternatively, build temporary executable and run:

```sh
go run ./cmd/server
```

## Example requests

Add user:

```sh
curl localhost:10002/v1/users -X POST -d '{"name": "Joe", "email": "example@domain.org"}'
```

Output:

```json
{"id":"4ea7e667-88aa-4dad-9e39-772ec8848da5"}
```

List users:

```sh
curl localhost:10002/v1/users -X GET
```

Output:

```json
{"users":[{"id":"4ea7e667-88aa-4dad-9e39-772ec8848da5","name":"Joe","email":"example@domain.org"}]}
```
