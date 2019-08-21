vcl 4.1;

backend default {
  .host = "127.0.0.1";
  .port = "3030";
}

sub vcl_recv {
  if (req.url ~ "^/api/flux") {
    set req.backend_hint = default;
  }
}

sub vcl_backend_response {
  set beresp.ttl = 5s;
  set beresp.grace = 30s;
}
