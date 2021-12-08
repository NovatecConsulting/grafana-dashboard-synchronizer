# Grafana Dashboard Synchronization Backend Plugin

A Grafana backend plugin for automatic synchronization of dashboard between multiple Grafana instances.

This plugin can be used to synchronize dashboards via a Git repository.  
A possible use case is: One Grafana instance is using the plugin to push dashboards to a Git repository,
another Grafana instance is using it to pull the dashboards.  
As an example this is useful to stage dashboards from "dev" to "prod" environments.

## Getting started

A data source backend plugin consists of both frontend and backend components.

### Frontend

1. Install dependencies

   ```bash
   yarn install
   ```

2. Build plugin in development mode or run in watch mode

   ```bash
   yarn dev
   ```

   or

   ```bash
   yarn watch
   ```

3. Build plugin in production mode

   ```bash
   yarn build
   ```

### Backend

1. Update [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/) dependency to the latest minor version:

   ```bash
   go get -u github.com/grafana/grafana-plugin-sdk-go
   go mod tidy
   ```

2. Build backend plugin binaries for Linux, Windows and Darwin:

   ```bash
   mage -v
   ```

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```

### Local Development

Set environment variables
```
PLUGIN_REPO = local path to cloned repo
GIT_SSH_KEY = path to private git sshkey
```

Build frontend and backend and start docker-compose

```
docker-compose up 
```

Under datasources the Grafana Dashboard Plugin Sync should be available now

## Releasing the Plugin

The release process of the plugin is automated using Github Actions.
On each push to the `main` branch, a new prerelease is created and the corresponding commit is tagged "latest".
Old prereleases will be deleted.

To create a normal release, the commit that is used as the basis for the release must be tagged with the following format: `v*.*.*`.
After that, the release is built and created with the version number extracted from the tag.
Furthermore, a new commit is created, which sets the current version in the `main` branch to the version that has been released.

## Learn more

- [Build a data source backend plugin tutorial](https://grafana.com/tutorials/build-a-data-source-backend-plugin)
- [Grafana documentation](https://grafana.com/docs/)
- [Grafana Tutorials](https://grafana.com/tutorials/) - Grafana Tutorials are step-by-step guides that help you make the most of Grafana
- [Grafana UI Library](https://developers.grafana.com/ui) - UI components to help you build interfaces using Grafana Design System
- [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/)
