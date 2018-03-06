FROM busybox

COPY main /

ENTRYPOINT ["/main"]
