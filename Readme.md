# 🌊 Stream-http

- Go-webserver wrapping the `exiftool -listx` command.
- Parses a subset of the `XML` output into `JSON`.
- Returns the result via the `/tags` endpoint on port `8080`.
  
## 🔎 Observations

- `exiftool` needs to be installed.
- The command's output is streamed to the output of the `/tags` endpint _while_ the command is still running.
- Parsing from `XML` to `JSON` is done by chunks of pairing `<table></table>` tags.
- Parsed chunks are appended continuously to the output of the `/tags` endpoint, in `JSON` format.
- The long running task can be cancelled, e.g. via the browser's stop button or issuing a sigterm via `CTRL-C` on the terminal.
  - The standard output pipe to the `exixftool` command call will be closed.
  - The command itself will be killed by the shared `Context()` object.

## 🚀 Run

> `go run main.go models.go`.

## 🤲🏽 Helpers

See [`cmd\curl_get_tags_unbuffered.sh`](https://github.com/levelont/stream-http/blob/main/cmd/curl_get_tags_unbuffered.sh) for an example on how to call the endpoint from the terminal.

## 🙏🏽 Thanks!
~Luis