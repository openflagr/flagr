# flagr-ui

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

### Compiles and minifies for production
```
npm run build
```

### Run your tests
```
npm run test
```

### Lints and fixes files
```
npm run lint
```

### Upgrade vue-cli
```
npm install -g @vue/cli
vue --version
vue upgrade
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

## Running Flagr Locally

1. Update Dockerfile (follow comments in Dockerfile):
    - Change port from 18000 to 3000
    - set FLAGR_RECORDER_ENABLED to false for local testing if you dont want kafka setup
    - Add JWT secret and
    - ENV FLAGR_JWT_AUTH_NO_TOKEN_REDIRECT_URL="http://localhost:3000/login"

2. Update frontend configuration:
   ```javascript
   // browser/src/constants.js
   DEV: {
       VUE_APP_API_URL: 'http://localhost:3000/api/v1',
       VUE_APP_SSO_API_URL: 'https://bff.allen-stage.in/internal-bff/',
   }
   ```

3. Build and run:
   ```bash
   # Build the Docker image
   docker build -t flagr .
 
   # Run the container
   docker run -it -p 3000:3000 flagr
   ```
