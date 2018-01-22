# pg\_elastic - ElasticSearch API frontend for PostgreSQL

[![Go Report Card](https://goreportcard.com/badge/github.com/asp437/pg_elastic)](https://goreportcard.com/report/github.com/asp437/pg_elastic)

`pg_elastic` application provides *ElasticSearch* API for storing and processing data in *PostgreSQL* RDBMS.

> **_IMPORTANT!_** `pg_elastic` requires Go compiler version 1.9 or above.

## Motivation

The main purpose of the project is to validate the idea to provide *ElasticSearch* API for users who want
to use *PostgreSQL* as a backend. Project main goal is to check is it possible to use *PostgreSQL* as a useful alternative
for *ElasticSearch* instance.

## Installation

`pg_elastic` written in *Go* programing language which is required to build the application.
All dependencies could be automatically downloaded via `go get` command.

To download source code with all dependcies into `$GOPATH` environment use:

```
go get github.com/asp437/pg_elastic
```

To build the application and place executable file in current directory use:

```
go build github.com/asp437/pg_elastic
```

To build and install the application int `$GOPATH/bin` use:

```
go install github.com/asp437/pg_elastic
```

Built executable file could be found at `$GOPATH/bin/pg_elastic`

To run dev-version of application inside its directory execute:

```
go run main.go
```

## Configuration

The configuration of the application is read from file `pg_elastic_config.json` in the working directory
during the starting phase. Configuration consists of following parameters:

* `ServerPort` - port to access `pg_elastic`. *ElasticSearch* API will be available at this TCP port.
* `DBServerAddress` - address and port for target *PostgreSQL* server. Should be written in `host:port` format.
* `DBLogin` - login to access PostgreSQL database.
* `DBPassword` - password used to access *PostgreSQL* database.
* `DBName` - name of the database used as storage for `pg_elastic`

An example of configuration file could be found in the repository.

## Usage

In order to use `pg_elastic`, you should configure it and run an executable file.
After running the application, you can use *ElasticSearch* API to manipulate data.

> **_CAUTION!_** Manual editing of data generated/used by `pg_elastic` could lead to unstable work or loss of data.

### Supported API

* `GET` `/_cluster/health` - Get health of the cluster. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-health.html)
* `GET` `/_bulk` - Perform a number of bulk operations. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html)
* `PUT` `/{index}/_mapping/{type}` - Put mapping for a type. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-put-mapping.html)
* `PUT` `/{index}` - Create index. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html)
* `HEAD` `/{index}` - Check index existance. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-exists.html)
* `GET` `/{index_wildcard}/_search` - Search for a document in index. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html)
* `GET` `/{index_wildcard}/{type_wildcard}/_search` - Search for a document with specified index and type. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/search.html)
* `PUT/POST` `/{index_wildcard}/{type_wildcard}/{id?}` - Insert a document. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html)
* `GET` `/{index_wildcard}/{type_wildcard}/{id}` - Get document with specified ID. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-get.html)
* `DELETE` `/{index_wildcard}/{type_wildcard}/{id}` - Delete document with specified ID. [Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-delete.html)

## Migration

To migrate data from existing *ElasticSearch* cluster into *PostgreSQL* instance for further usage with `pg_elastic`
use migration tool which is available in the `pg_elastic_migrate` subpackage. It is automatically downloaded together with
`pg_elastic` itself.

To build the application and place executable file in current directory use:

```
go build github.com/asp437/pg_elastic/pg_elastic_migrate
```

To build and install it into `$GOPATH/bin` use:

```
go install github.com/asp437/pg_elastic/pg_elastic_migrate
```

To use the tool you should provide information about both *ElasticSearch*
source server and *PostgreSQL* destanation instance. For example(all parameters are optional):

```
pg_elastic_migrate -elasticsearch-host=localhost:9200 -postgresql-host=localhost:5432 -postgresql-user=postgres -postgresql-password=postgres -postgresql-database=postgres
```

Built executable file could be found at `$GOPATH/bin/pg_elastic_migrate`

Use `-h` flag to get more information about tool's parameters.

> **_CAUTION!_** A building of `pg_elastic_migrate` inside of `pg_elastic` directory via `go build` command
> leads to error caused by naming collision. Please use another directory for manual building or use automatic install
> instead.

## Testing

Directory tests contains black-box testing of provided API. It requires *Python* interpreter with a number
of dependecies. All dependencies can be downloaded via `virtualenv` based on content of `requirements.txt`

```
pip3 install -r requirements.txt
```

Tests are based on *Pytest* framework. In order to run tests execute the following command:

```
pytest
```

Tests can test both `pg_elastic` or *ElasticSearch* to guarantee the correctness of the tests.
You can specify TCP port for testing target via `ELASTIC_PORT` environment variable.

```
ELASTIC_PORT=9200 pytest
```

The default port is `9200` which is a standard port for *ElasticSearch* installation.

## Further Development

Currently, the project is frozen in a stage of providing only basic API. For further development an implementation
of different parts of *ElasticSearch* is required, e.g. *PostgreSQL* dictionaries, parsers, changes in
`pg_elastic` itself, etc.
