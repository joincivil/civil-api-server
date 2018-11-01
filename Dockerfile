#FROM alpine:3.7
FROM golang:1.11.1
ADD build build
ADD build/graphqlserver /graphqlserver
RUN chmod u+x /graphqlserver

CMD ["/graphqlserver", "-logtostderr=true", "-stderrthreshold=INFO"]

