FROM alpine:latest
ENV NODE ""
WORKDIR /app
RUN apk update && apk add --no-cache tzdata
RUN cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime
RUN echo "Europe/Berlin" >  /etc/timezone
COPY nubedb .
RUN mkdir /app/data
RUN touch /app/data/.keep
ENTRYPOINT ["/app/nubedb"]
