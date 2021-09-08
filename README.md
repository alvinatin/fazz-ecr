# fazz-ecr

Tools for interacting with the ECR Docker registry
`322727087874.dkr.ecr.ap-southeast-1.amazonaws.com`.

This repo provide two utilities:

- `docker-credential-fazz-ecr`: Credential helper for Docker client.
- `fazz-ecr-create-repo`: Helper to create new repository on the
  `322727087874.dkr.ecr.ap-southeast-1.amazonaws.com` registry.

User permissions to repositories are determined by membership of Google groups.

## How to install

Run `go install github.com/payfazz/fazz-ecr/cmd/...@latest` to install both
utilities in your `$GOBIN`.

## Quickstart

- Run `docker-credential-fazz-ecr update-config`. If `{HOME}/.docker/config.json`
  doesn't exist, create the file with `{}` (this is empty JSON object) content,
  and then run the command again.

- Create a new repository with command
  `fazz-ecr-create-repo 322727087874.dkr.ecr.ap-southeast-1.amazonaws.com/{owner}/{repository_name}`,
  where `{owner}` is your email or group name but all characters outside the
  `a-zA-Z0-9-_` regex is replaced with `-`. For example, if your email is
  `win@payfazz.com`, you can create a repository with this command
  `fazz-ecr-create-repo 322727087874.dkr.ecr.ap-southeast-1.amazonaws.com/win-payfazz-com/example-service`.

## How to use in GitHub Actions

Use `payfazz/setup-fazz-ecr-action@v1` action in your workflow file. Because CI
environment is not interactive, `FAZZ_ECR_TOKEN` environment variable must be
set.

# `docker-credential-fazz-ecr`

`docker-credential-fazz-ecr` provides sub-command:

- `update-config`
- `login`
- `list-access`
- `get`

`docker-credential-fazz-ecr update-config` is used to update
`~/.docker/config.json` so that any access to registry
`322727087874.dkr.ecr.ap-southeast-1.amazonaws.com` is using `fazz-ecr`
credential

`docker-credential-fazz-ecr login` is used to remove old credential and retrieve
new credential, this command will open web browser to authenticate your identity

`docker-credential-fazz-ecr list-access` is used to print which repository is
accessible by current credentials

`docker-credential-fazz-ecr get` is used internally by `docker` command line
