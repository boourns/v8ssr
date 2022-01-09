# v8ssr

Embedded javascript server-side renderer for Golang.

Useful for static server-side rendering.  This does not attempt to polyfill node or browser features into the v8 engine, so it may not work for isomorphic rendering if features are expected to exist (like fetch, or document).  It is designed to server-side render a read-only, public-facing typescript/preact frontend.

Uses [v8go](https://github.com/rogchap/v8go)

## Features
- Pass any object from Go into the JS context (the object is JSON marshalled when crossing the bridge)
- Multi-threaded, run a configurable number of threads
- Caches the loaded script
- Supports reloading bundle from disk when it changes (avoid restarting server in development)

## Acknowledgements

Loosely based on https://github.com/tmc/reactssr
