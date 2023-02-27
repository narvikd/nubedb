FROM alpine:latest
ENV FSM_PERFORMANCE ""
WORKDIR /app
COPY nubedb .
RUN mkdir /app/data
RUN touch /app/data/.keep
ENTRYPOINT ["/app/nubedb"]
