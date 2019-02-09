## Setup

```sh
$ go mod download
```

## Test and Lint

```sh
$ make lint
$ make test
```

## Cloud Build
```sh
$ gcloud builds submit --config cloudbuild.yaml ./
```
