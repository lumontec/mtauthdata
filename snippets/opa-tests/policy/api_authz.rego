package httpapi.authz

# HTTP API request
import input

default allow = false


# Allow to access the /docs route.
allow {
  input.path = ["/docs"]
  input.method = "GET"
}

# Allow to access the /docs/{doc_id}/value route.
allow {
  input.path = ["/docs", doc_id, "/value"]
  input.method = "POST"
}

