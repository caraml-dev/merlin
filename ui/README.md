# Working with the Merlin UI

This file explains how to work with the React-based Merlin UI.

## Introduction

The [React-based](https://reactjs.org/) Merlin UI was bootstrapped using [Create React App](https://github.com/facebook/create-react-app), a popular toolkit for generating React application setups. You can find general information about Create React App on [their documentation site](https://create-react-app.dev/).

## Development environment

To work with the React UI code, you will need to have the following tools installed:

- The [Node.js](https://nodejs.org/) JavaScript runtime.
- The [Yarn](https://yarnpkg.com/) package manager.

## Installing npm dependencies

The React UI depends on a large number of [npm](https://www.npmjs.com/) packages. These are not checked in, so you will need to download and install them locally via the Yarn package manager:

    yarn

Yarn consults the `package.json` and `yarn.lock` files for dependencies to install. It creates a `node_modules` directory with all installed dependencies.

## Running a local development server

You need to set the `REACT_APP_OAUTH_CLIENT_ID` in `.env.development` with the Google Oauth Client ID for authentication.

Then, you can start a development server for the React UI outside of a running Merlin server by running:

    yarn start

This will open a browser window with the React app running on http://localhost:3000/. The page will reload if you make edits to the source code. You will also see any lint errors in the console.

Due to a `"proxy": "http://localhost:8080/v1"` setting in the `package.json` file, any API requests from the React UI are proxied to `localhost` on port `8080` by the development server. This allows you to run a normal Merlin server to handle API requests, while iterating separately on the UI.

    [browser] ----> [localhost:3000 (dev server)] --(proxy API requests)--> [localhost:8080/v1 (Merlin)]

## Linting

We use [lint-staged](https://github.com/okonet/lint-staged) for the linter. To detect and automatically fix lint errors against staged git files, run:

    yarn lint

This is also available via the `lint-ui` target in the main Merlin `Makefile`.

## Building the app for production

To build a production-optimized version of the React app to a `build` subdirectory, run:

    yarn build

**NOTE:** You will likely not need to do this directly. Instead, this is taken care of by the `build` target in the main Merlin `Makefile` when building the full binary.

## Integration into Merlin

To build a Merlin binary that includes a compiled-in version of the production build of the React app, change to the root of the repository and run:

    make build

This installs npm dependencies via Yarn, builds a production build of the React app, and then finally compiles in all web assets into the Merlin binary.
