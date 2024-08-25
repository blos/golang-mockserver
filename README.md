# Mockserver

This project is a fun project to get familiar with Go. Also this could help mock dependencies to services.
You define a route and a wanted static response and start the server.
Go to the defined URL will yield the defined response.

## Run or build it
Use the Makefile by using
```shell
make run
```
or
```shell
make build
```

## Configs
There are 2 configurations that can be made for this mockserver: Config the server itself and config the responses that you want to mock.
### Server configs
These configs can be set by environment variable or by CLI parameter (env. var/CLI param.):
- `MOCK_SERVER_HOST`/`--host`: Host address the server listens to (default "0.0.0.0")
- `MOCK_SERVER_PORT`/`--port`: Port the server listens to (default "1080")
- `MOCK_ROUTES_FILE`/`--routes-file`: Path to the route definitions (default "./example.routes.json")
- `--verbose`: Makes the logging of the mockserver verbose
### Server response configs
Consider the file `exmaple.routes.json` for an example.

The responses are configured by single a JSON file.
For each route you can configure the following entries:
- `path` (Required): A string for the URL path, e.g. `http://example.com/v1` in which `/v1` is the path.
- `method` (Required): The HTTP method that you want to speak to. Currently the following are implemented: `GET`, `PUT`, `POST`, `PATCH`, `DELETE`
- `status_code`: The status code that is returned for a mocked path.
- `response_offset_mode`: Can be "constant" or "normal". For "constant" always the same time is used to await until the response is send. "Normal" uses an normal distribution, so that every response has a different response time.
- `response_offset`: If `response_offset_mode="constant"` this must be an integer respresenting the waiting time in milliseconds. Example
```JSON
...
"response_offset_mode": "constant",
"response_offset": 1000, // = 1 second
...
```
If `response_offset_mode="normal"` this is an object with entries "mean" and "std" for the normal distribution.
```JSON
...
"response_offset_mode": "normal",
"response_offset": {
    "mean": 2000, // = 2 seconds
    "std": 1000   // = 1 second
},
...
```
- `header`: An object of header entries that should be returned in the mocked response such as
```JSON
...
"header": {
    "X-KEY1": "VAL1",
    "X-KEY2": "VAL2",
},
...
```
- `body`: The body of the response, also known as payload:
```JSON
...
"body": {
    "test-entry-key": "test-entry-value"
},
...
```

## Things to come
- [x] Response time offset: Response is sent after a constant offset.
- [x] Normal distributed offset behavior: The offset is chosen randomly when the request is made
- [X] Update documentation: Insert examples and explainations in the README
- [x] Add CLI parameters and help
- [x] Add verbose option for the mockserver
- [ ] Add tests
- [ ] ...