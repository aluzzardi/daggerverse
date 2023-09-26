# Alpine module

This module provides utilities for alpine-based containers.

## Example

```console
$ cat example.graphql
{
  alpine {
    withVersion(version: "3.18.2") {
      withPackage(name: "curl") {
        container {
          withExec(args: ["curl", "https://dagger.io"]) {
            stdout
          }
        }
      }
    }
  }
}

$ cat example.graphql | dagger query
...
```
