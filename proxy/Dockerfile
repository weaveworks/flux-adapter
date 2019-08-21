FROM varnish:6.2

CMD ["varnishd", "-F", "-a", "127.0.0.1:3033", "-T", "none", "-f", "/etc/varnish/default.vcl"]

COPY default.vcl /etc/varnish/default.vcl

