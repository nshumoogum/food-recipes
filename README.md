food-recipes
=============

### Getting started

The food recipe API can be run by running `make debug`.

#### Import data from google sheets

#### Configuration

| Environment variable         | Default                                | Description
| ---------------------------- | ---------------------------------------| -----------
| BIND_ADDR                    | :30000                                 | The host and port to bind to
| CONNECTION_STRING            | ""                                     | Unique key to allow access to write endpoints. Should be set to something
| DOWNLOAD_DATA                | false                                  | Flag to determine whether to attempt to download recipes from google sheet
| DOWNLOAD_TIMEOUT             | 5s                                     | The download google sheet timeout in seconds
| GOOGLE_SHEET_URL             | ""                                     | The published url for the google sheet containing recipes 
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                                     | The graceful shutdown timeout in seconds

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Released under MIT license, see [LICENSE](LICENSE) for details.
