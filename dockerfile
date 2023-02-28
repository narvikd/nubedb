FROM alpine:latest
ENV FSM_PERFORMANCE ""
ENV DISCOVER_STRATEGY ""
WORKDIR /app
COPY nubedb .
RUN mkdir /app/data
RUN touch /app/data/.keep
ENTRYPOINT ["/app/nubedb"]
