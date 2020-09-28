## Getting Started

Install all dependencies:

### `npm install`

Initialize environment variables:

```
export REACT_APP_API=http://localhost:3000/v1
export REACT_APP_OAUTH_CLIENT_ID=<oauth_client_id>
```

REACT_APP_API should points to merlin API and REACT_APP_OAUTH_CLIENT_ID is the client ID used for authentication.

And you can run:

### `npm start`

Runs the app in the development mode.<br>
Open [http://localhost:3001](http://localhost:3001) to view it in the browser.

The page will reload if you make edits.<br>

You can also use mock data by setting `USE_MOCK_DATA: true` in [config.js](src/config.js)  
All mock data is configured under [mocks](src/mocks) folder


You will also see any lint errors in the console.

### `npm test`

Launches the test runner in the interactive watch mode.<br>
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `npm run build`

Builds the app for production to the `build` folder.<br>
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.<br>
Your app is ready to be deployed!