# fazz-ecr
Tooling for accessing ecr docker registry `322727087874.dkr.ecr.ap-southeast-1.amazonaws.com`

this repo provide 2 utility, `docker-credential-fazz-ecr` docker credential helper, and `fazz-ecr-create-repo` to create repository

access permission is managed by google groups

## how to install
run `go go install github.com/payfazz/fazz-ecr/cmd/...@latest`, this will install `docker-credential-fazz-ecr` and `fazz-ecr-create-repo` in your GOBIN directory

run `docker-credential-fazz-ecr update-config`, this will update your `~/docker/config.json`

## how to use in github action
use `payfazz/setup-fazz-ecr-action@v1` action

because CI environment is not interactive, you must provide `FAZZ_ECR_TOKEN` env
