# starter-app

a starter app with all the things you need for a new microservice:

- web server
- .env file parsing
- database
- cli for doing up/down db migrations

## installation

1. install [asdf](https://github.com/asdf-vm/asdf)
2. install [asdf plugins](https://github.com/asdf-vm/asdf-plugins), see `.tool-versions` for plugins
3. install tool versions listed in `.tool-versions`

## running

1. ensure sure database is running, see `docker-compose up --force-recreate`
2. apply any db migrations, see `go run app/cmd/migrate/main.go -h`
3. run web server, see `mage run:server`
