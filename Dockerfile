# Multi-stage docker build
# Build stage
FROM golang:alpine

COPY taskEngineFunctionalCheck taskEngineFunctionalCheck

ENTRYPOINT ["./taskEngineFunctionalCheck"]